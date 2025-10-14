package oxc

/*
#cgo LDFLAGS: -L${SRCDIR}/../../oxc-resolver-ffi/target/release -loxc_resolver_ffi
#cgo darwin LDFLAGS: -framework CoreFoundation -framework Security
#include "../../oxc-resolver-ffi/oxc_resolver.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// Resolver wraps the oxc_resolver C library
type Resolver struct {
	ptr *C.CResolver
}

// ResolveOptions contains options for creating a resolver
type ResolveOptions struct {
	ConditionNames []string
	Extensions     []string
	MainFields     []string
	ExportsFields  []string
	TsconfigPath   string

	FullySpecified  bool
	Symlinks        bool
	PreferRelative  bool
	DeclarationOnly bool
}

// Resolution represents a resolved module
type Resolution struct {
	Path                     string
	Query                    string
	Fragment                 string
	ResolvedUsingTsExtension bool
}

// ResolveError represents an error during resolution
type ResolveError struct {
	Code    int
	Message string
}

func (e *ResolveError) Error() string {
	return fmt.Sprintf("resolve error (code %d): %s", e.Code, e.Message)
}

// NewResolver creates a new resolver with the given options
func NewResolver(options *ResolveOptions) (*Resolver, error) {
	if options == nil {
		options = &ResolveOptions{}
	}

	// Convert Go strings to C strings
	var cConditionNames **C.char
	var cExtensions **C.char
	var cMainFields **C.char
	var cExportsFields **C.char

	conditionNamesLen := len(options.ConditionNames)
	extensionsLen := len(options.Extensions)
	mainFieldsLen := len(options.MainFields)
	exportsFieldsLen := len(options.ExportsFields)

	// Track C strings to free them later
	var cStringsToFree []*C.char
	defer func() {
		for _, cs := range cStringsToFree {
			C.free(unsafe.Pointer(cs))
		}
	}()

	// Allocate C string arrays
	if conditionNamesLen > 0 {
		cConditionNames = (**C.char)(C.malloc(C.size_t(conditionNamesLen) * C.size_t(unsafe.Sizeof(uintptr(0)))))
		defer C.free(unsafe.Pointer(cConditionNames))
		for i, s := range options.ConditionNames {
			cs := C.CString(s)
			cStringsToFree = append(cStringsToFree, cs)
			*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cConditionNames)) + uintptr(i)*unsafe.Sizeof(uintptr(0)))) = cs
		}
	}

	if extensionsLen > 0 {
		cExtensions = (**C.char)(C.malloc(C.size_t(extensionsLen) * C.size_t(unsafe.Sizeof(uintptr(0)))))
		defer C.free(unsafe.Pointer(cExtensions))
		for i, s := range options.Extensions {
			cs := C.CString(s)
			cStringsToFree = append(cStringsToFree, cs)
			*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cExtensions)) + uintptr(i)*unsafe.Sizeof(uintptr(0)))) = cs
		}
	}

	if mainFieldsLen > 0 {
		cMainFields = (**C.char)(C.malloc(C.size_t(mainFieldsLen) * C.size_t(unsafe.Sizeof(uintptr(0)))))
		defer C.free(unsafe.Pointer(cMainFields))
		for i, s := range options.MainFields {
			cs := C.CString(s)
			cStringsToFree = append(cStringsToFree, cs)
			*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cMainFields)) + uintptr(i)*unsafe.Sizeof(uintptr(0)))) = cs
		}
	}

	if exportsFieldsLen > 0 {
		cExportsFields = (**C.char)(C.malloc(C.size_t(exportsFieldsLen) * C.size_t(unsafe.Sizeof(uintptr(0)))))
		defer C.free(unsafe.Pointer(cExportsFields))
		for i, s := range options.ExportsFields {
			cs := C.CString(s)
			cStringsToFree = append(cStringsToFree, cs)
			*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cExportsFields)) + uintptr(i)*unsafe.Sizeof(uintptr(0)))) = cs
		}
	}

	// Handle tsconfig path
	var cTsconfigPath *C.char = nil
	if options.TsconfigPath != "" {
		cTsconfigPath = C.CString(options.TsconfigPath)
		defer C.free(unsafe.Pointer(cTsconfigPath))
	}

	// Create C options struct with explicit zero values for all fields
	cOpts := C.CResolveOptions{
		condition_names:     cConditionNames,
		condition_names_len: C.size_t(conditionNamesLen),
		extensions:          cExtensions,
		extensions_len:      C.size_t(extensionsLen),
		main_fields:         cMainFields,
		main_fields_len:     C.size_t(mainFieldsLen),
		exports_fields:      cExportsFields,
		exports_fields_len:  C.size_t(exportsFieldsLen),
		tsconfig_path:       cTsconfigPath,
		enforce_extension:   C.bool(false),
		fully_specified:     C.bool(options.FullySpecified),
		symlinks:            C.bool(options.Symlinks),
		prefer_relative:     C.bool(options.PreferRelative),
		declaration_only:    C.bool(options.DeclarationOnly),
		_reserved:           [8]C.uint64_t{0, 0, 0, 0, 0, 0, 0, 0},
	}

	// Create resolver
	ptr := C.oxc_resolver_new(&cOpts)
	if ptr == nil {
		return nil, errors.New("failed to create resolver")
	}

	resolver := &Resolver{ptr: ptr}
	runtime.SetFinalizer(resolver, (*Resolver).free)

	return resolver, nil
}

// Resolve resolves a module specifier from the given directory
func (r *Resolver) Resolve(path, specifier string) (*Resolution, error) {
	if r.ptr == nil {
		return nil, errors.New("resolver is nil")
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	cSpecifier := C.CString(specifier)
	defer C.free(unsafe.Pointer(cSpecifier))

	// Call C function
	cRes := C.oxc_resolver_resolve(r.ptr, cPath, cSpecifier)

	// Always free C strings
	defer C.oxc_resolution_free(&cRes)

	// Check for error
	if cRes.error_code != 0 {
		var msg string
		if cRes.error_message != nil {
			msg = C.GoString(cRes.error_message)
		} else {
			msg = "unknown error"
		}
		return nil, &ResolveError{
			Code:    int(cRes.error_code),
			Message: msg,
		}
	}

	// Convert result to Go
	res := &Resolution{
		ResolvedUsingTsExtension: bool(cRes.resolved_using_ts_extension),
	}

	if cRes.path != nil {
		res.Path = C.GoString(cRes.path)
	}
	if cRes.query != nil {
		res.Query = C.GoString(cRes.query)
	}
	if cRes.fragment != nil {
		res.Fragment = C.GoString(cRes.fragment)
	}

	return res, nil
}

// ResolveTypeReferenceDirective resolves a type reference directive
func (r *Resolver) ResolveTypeReferenceDirective(containingFile, typeReference string) (*Resolution, error) {
	if r.ptr == nil {
		return nil, errors.New("resolver is nil")
	}

	cContainingFile := C.CString(containingFile)
	defer C.free(unsafe.Pointer(cContainingFile))

	cTypeReference := C.CString(typeReference)
	defer C.free(unsafe.Pointer(cTypeReference))

	// Call C function
	cRes := C.oxc_resolver_resolve_type_reference_directive(r.ptr, cContainingFile, cTypeReference)

	// Always free C strings
	defer C.oxc_resolution_free(&cRes)

	// Check for error
	if cRes.error_code != 0 {
		var msg string
		if cRes.error_message != nil {
			msg = C.GoString(cRes.error_message)
		} else {
			msg = "unknown error"
		}
		return nil, &ResolveError{
			Code:    int(cRes.error_code),
			Message: msg,
		}
	}

	// Convert result to Go
	res := &Resolution{
		ResolvedUsingTsExtension: bool(cRes.resolved_using_ts_extension),
	}

	if cRes.path != nil {
		res.Path = C.GoString(cRes.path)
	}
	if cRes.query != nil {
		res.Query = C.GoString(cRes.query)
	}
	if cRes.fragment != nil {
		res.Fragment = C.GoString(cRes.fragment)
	}

	return res, nil
}

// free releases the resolver resources
func (r *Resolver) free() {
	if r.ptr != nil {
		C.oxc_resolver_free(r.ptr)
		r.ptr = nil
	}
}

// Free explicitly frees the resolver (optional, called automatically by GC)
func (r *Resolver) Free() {
	r.free()
	runtime.SetFinalizer(r, nil)
}
