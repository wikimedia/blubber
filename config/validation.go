package config

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/distribution/distribution/reference"
	"gopkg.in/go-playground/validator.v9"
)

var (
	// See Debian Policy
	//  https://www.debian.org/doc/debian-policy/#s-f-source
	//  https://www.debian.org/doc/debian-policy/#s-f-version
	debianPackageName   = `[a-z0-9][a-z0-9+.\-]+`
	debianVersionSpec   = `(?:[0-9]+:)?[0-9]+[a-zA-Z0-9\.\+\-~]*`
	debianReleaseName   = `[a-zA-Z](?:[a-zA-Z0-9\-]*[a-zA-Z0-9]+)?`
	debianComponent     = `[a-z0-9][a-z0-9+./\-]+`
	debianReleaseRegexp = regexp.MustCompile(debianReleaseName)
	debianPackageRegexp = regexp.MustCompile(fmt.Sprintf(
		`^%s(?:=%s|/%s)?$`, debianPackageName, debianVersionSpec, debianReleaseName))
	debianComponentRegexp = regexp.MustCompile(debianComponent)

	// See IEEE Std 1003.1-2008 (http://pubs.opengroup.org/onlinepubs/9699919799/)
	environmentVariableRegexp = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]+$`)

	// Pattern for valid variant names
	variantNameRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\-\.]+[a-zA-Z0-9]$`)

	// Pattern for Python package version constraints. We allow a subset of
	// the spec that omits support for extras, environment markers, and urls.
	// See https://www.python.org/dev/peps/pep-0508/#specification
	pythonVersionCmp      = `(?:<|<=|!=|==|>=|>|~=)`
	pythonVersion         = `[a-z-A-Z0-9\-_\.\*\+!]+`
	pythonVersionOne      = fmt.Sprintf(`%s%s`, pythonVersionCmp, pythonVersion)
	pythonContraintRegexp = regexp.MustCompile(fmt.Sprintf(
		`^%s(?:,%s)*$`, pythonVersionOne, pythonVersionOne))

	humanizedErrors = map[string]string{
		"abspath":           `{{.Field}}: "{{.Value}}" is not a valid absolute non-root path`,
		"artifactfrom":      `{{.Field}}: "{{.Value}}" is not a valid image reference or known variant`,
		"currentversion":    `{{.Field}}: config version "{{.Value}}" is unsupported`,
		"debiancomponent":   `{{.Field}}: "{{.Value}}" is not a valid Debian component name`,
		"debianpackage":     `{{.Field}}: "{{.Value}}" is not a valid Debian package name`,
		"debianrelease":     `{{.Field}}: "{{.Value}}" is not a valid Debian release name`,
		"envvars":           `{{.Field}}: contains invalid environment variable names`,
		"httpurl":           `{{.Field}}: "{{.Value}}" is not a valid HTTP/HTTPS URL`,
		"imageref":          `{{.Field}}: "{{.Value}}" is not a valid image reference`,
		"nodeenv":           `{{.Field}}: "{{.Value}}" is not a valid Node environment name`,
		"pypkgver":          `{{.Field}}: "{{.Value}}" is not a valid Python package version specification`,
		"relativelocal":     `{{.Field}}: path must be relative when "from" is "local"`,
		"required":          `{{.Field}}: is required`,
		"requiredwith":      `{{.Field}}: is required if "{{.Param}}" is also set`,
		"unique":            `{{.Field}}: cannot contain duplicates`,
		"uniqueartifacts":   `{{.Field}}: cannot contain duplicates`,
		"username":          `{{.Field}}: "{{.Value}}" is not a valid user name`,
		"variantref":        `{{.Field}}: references an unknown variant "{{.Value}}"`,
		"variants":          `{{.Field}}: contains a bad variant name`,
		"uniquetypesexcept": `{{.Field}}: contains disallowed repetitions of entries`,
		"notallowedwith":    `{{.Field}}: is not allowed if any of field(s) "{{.Param}}" is declared/included`,
	}

	validatorAliases = map[string]string{
		"currentversion": "eq=" + CurrentVersion,
		"nodeenv":        "alphanum",
		"username":       "hostname,ne=root",
		"artifactfrom":   "variantref|imageref",
	}

	validatorFuncs = map[string]validator.FuncCtx{
		"abspath":         isAbsNonRootPath,
		"debiancomponent": isDebianComponent,
		"debianpackage":   isDebianPackage,
		"debianrelease":   isDebianRelease,
		"envvars":         isEnvironmentVariables,
		"httpurl":         isHTTPURL,
		"imageref":        isImageRef,
		"isfalse":         isFalse,
		"istrue":          isTrue,
		"pypkgver":        isPythonPackageVersion,
		"relativelocal":   isRelativePathForLocalArtifact,
		"requiredwith":    isSetIfOtherFieldIsSet,
		"uniqueartifacts": uniqueByEquality[ArtifactsConfig],
		"variantref":      isVariantReference,
		"variants":        hasVariantNames,
		// Can be used with array or slice fields
		"uniquetypesexcept": uniqueTypesExcept,
		// Note this validator cannot be used with fields that contain a struct. See function
		// `notAllowedWith` for more details
		"notallowedwith": notAllowedWith,
	}
)

type ctxKey uint8

const rootCfgCtx ctxKey = iota

// newValidator returns a validator instance for which our custom aliases and
// functions are registered.
func newValidator() *validator.Validate {
	validate := validator.New()

	validate.RegisterTagNameFunc(resolveJSONTagName)

	for name, tags := range validatorAliases {
		validate.RegisterAlias(name, tags)
	}

	for name, f := range validatorFuncs {
		validate.RegisterValidationCtx(name, f, true)
	}

	return validate
}

// Validate runs all validations defined for config fields against the given
// Config value. If the returned error is not nil, it will contain a
// user-friendly message describing all invalid field values.
func Validate(config interface{}) error {
	validate := newValidator()

	ctx := context.WithValue(context.Background(), rootCfgCtx, config)

	return validate.StructCtx(ctx, config)
}

// HumanizeValidationError transforms the given validator.ValidationErrors
// into messages more likely to be understood by human beings.
func HumanizeValidationError(err error) string {
	var message bytes.Buffer

	if err == nil {
		return ""
	} else if !IsValidationError(err) {
		return err.Error()
	}

	templates := map[string]*template.Template{}

	for name, tmplString := range humanizedErrors {
		if tmpl, err := template.New(name).Parse(tmplString); err == nil {
			templates[name] = tmpl
		}
	}

	for _, ferr := range err.(validator.ValidationErrors) {
		if tmpl, ok := templates[ferr.Tag()]; ok {
			tmpl.Execute(&message, ferr)
		} else if trueErr, ok := err.(error); ok {
			message.WriteString(trueErr.Error())
		}

		message.WriteString("\n")
	}

	return strings.TrimSpace(message.String())
}

// IsValidationError tests whether the given error is a
// validator.ValidationErrors and can be safely iterated over as such.
func IsValidationError(err error) bool {
	if err == nil {
		return false
	} else if _, ok := err.(*validator.InvalidValidationError); ok {
		return false
	} else if _, ok := err.(validator.ValidationErrors); ok {
		return true
	}

	return false
}

func hasVariantNames(_ context.Context, fl validator.FieldLevel) bool {
	for _, name := range fl.Field().MapKeys() {
		if !variantNameRegexp.MatchString(name.String()) {
			return false
		}
	}

	return true
}

func isAbsNonRootPath(_ context.Context, fl validator.FieldLevel) bool {
	value := fl.Field().String()

	return path.IsAbs(value) && path.Base(path.Clean(value)) != "/"
}

func isDebianComponent(_ context.Context, fl validator.FieldLevel) bool {
	value := fl.Field().String()

	return debianComponentRegexp.MatchString(value)
}

func isDebianPackage(_ context.Context, fl validator.FieldLevel) bool {
	value := fl.Field().String()

	return debianPackageRegexp.MatchString(value)
}

func isDebianRelease(_ context.Context, fl validator.FieldLevel) bool {
	value := fl.Field().String()

	return debianReleaseRegexp.MatchString(value)
}

func isHTTPURL(_ context.Context, fl validator.FieldLevel) bool {
	url, err := url.Parse(fl.Field().String())

	if err != nil {
		return false
	}

	return url.Scheme == "http" || url.Scheme == "https"
}

func isImageRef(_ context.Context, fl validator.FieldLevel) bool {
	value := fl.Field().String()

	return reference.ReferenceRegexp.MatchString(value)
}

func isEnvironmentVariables(_ context.Context, fl validator.FieldLevel) bool {
	for _, key := range fl.Field().MapKeys() {
		if !environmentVariableRegexp.MatchString(key.String()) {
			return false
		}
	}

	return true
}

func isFalse(_ context.Context, fl validator.FieldLevel) bool {
	val, ok := fl.Field().Interface().(bool)

	return ok && val == false
}

func isPythonPackageVersion(_ context.Context, fl validator.FieldLevel) bool {
	value := fl.Field().String()

	return pythonContraintRegexp.MatchString(value)
}

func isRelativePathForLocalArtifact(_ context.Context, fl validator.FieldLevel) bool {
	value := fl.Field().String()
	from := fl.Parent().FieldByName("From").String()

	if value == "" || from != LocalArtifactKeyword {
		return true
	}

	// path must be relative and do no "../" funny business
	return !(path.IsAbs(value) || strings.HasPrefix(path.Clean(value), ".."))
}

func isSetIfOtherFieldIsSet(_ context.Context, fl validator.FieldLevel) bool {
	if otherField, err := ResolveJSONPath(fl.Param(), fl.Parent().Interface()); err == nil {
		return isZeroValue(reflect.ValueOf(otherField)) || !isZeroValue(fl.Field())
	}

	return false
}

func isTrue(_ context.Context, fl validator.FieldLevel) bool {
	val, ok := fl.Field().Interface().(bool)

	return ok && val == true
}

func isVariantReference(ctx context.Context, fl validator.FieldLevel) bool {
	cfg := ctx.Value(rootCfgCtx).(Config)
	ref := fl.Field().String()

	if ref == LocalArtifactKeyword {
		return true
	}

	for name := range cfg.Variants {
		if name == ref {
			return true
		}
	}

	return false
}

// Supports array and slice fields
func uniqueTypesExcept(_ context.Context, fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() != reflect.Array && field.Kind() != reflect.Slice {
		return true
	}

	presentTypes := make(map[string]interface{})
	for i := 0; i < field.Len(); i++ {
		actualType := reflect.TypeOf(field.Index(i).Interface()).String()
		if _, isPresent := presentTypes[actualType]; isPresent {
			exception := fl.Param()
			if !strings.Contains(actualType, exception) {
				return false
			}
		}
		presentTypes[actualType] = new(interface{})
	}

	return true
}

type equality[T any] interface {
	Equal(T) bool
}

// uniqueByEquality iterates over a slice field twice and ensures it has no
// duplicate items. If there is an instance of equality (based on the
// [equality] interface) between two entries at different index positions, the
// validation will fail.
//
// The complexity of this function is O((n^2)-n) and should only be used with
// slice fields whose member type is not [comparable] by Go rules (if the
// member type itself contains a slice field, it is not [comparable]).
func uniqueByEquality[T equality[E], E any](_ context.Context, fl validator.FieldLevel) bool {
	field := fl.Field()

	for i := 0; i < field.Len(); i++ {
		v1 := field.Index(i).Interface().(T)

		for j := 0; j < field.Len(); j++ {
			v2 := field.Index(j).Interface().(E)
			if i != j && v1.Equal(v2) {
				return false
			}
		}
	}

	return true
}

// Validator field parameters (fl.Param() call in this code) are not available to user-defined
// validators when registered as a struct or custom validator. That means this validator cannot be
// registered for structs and doesn't work for fields that contain one (gets ignored)
func notAllowedWith(_ context.Context, fl validator.FieldLevel) bool {
	field := fl.Field()

	if isNullable(field) && field.IsNil() {
		return true
	}

	switch field.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		if field.Len() == 0 {
			return true
		}
	}

	if field.Kind() == reflect.String && field.Len() == 0 {
		return true
	}

	disallowingFields := strings.Fields(fl.Param())
	for _, disallowingField := range disallowingFields {
		if otherField, found := fieldByJSONName(fl.Parent(), disallowingField); found {
			if !isZeroValue(otherField) {
				return false
			}
		}
	}

	return true
}

func isNullable(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func resolveJSONTagName(field reflect.StructField) string {
	return strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
}

var typeToJSONNameToFieldName = make(map[reflect.Type]map[string]string)
var fieldByJSONNameMu sync.Mutex

func fieldByJSONName(v reflect.Value, jsonName string) (reflect.Value, bool) {
	t := v.Type()

	fieldByJSONNameMu.Lock()
	defer fieldByJSONNameMu.Unlock()

	if jsonNameToFieldName := typeToJSONNameToFieldName[t]; jsonNameToFieldName != nil {
		if fieldName, ok := jsonNameToFieldName[jsonName]; ok {
			return v.FieldByName(fieldName), true
		}

		typeToJSONNameToFieldName[t][jsonName] = ""
	} else {
		typeToJSONNameToFieldName[t] = make(map[string]string)
	}

	for _, field := range reflect.VisibleFields(t) {
		name := resolveJSONTagName(field)
		typeToJSONNameToFieldName[t][name] = field.Name
		if name == jsonName {
			return v.FieldByName(field.Name), true
		}
	}

	return reflect.Value{}, false
}
