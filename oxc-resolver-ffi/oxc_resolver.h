#ifndef OXC_RESOLVER_H
#define OXC_RESOLVER_H

#pragma once

#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

/**
 * Opaque resolver handle
 */
typedef struct CResolver CResolver;

/**
 * Options for resolver creation
 */
typedef struct CResolveOptions {
  /**
   * NULL-terminated array of condition names (NULL-terminated strings)
   */
  const char *const *condition_names;
  uintptr_t condition_names_len;
  /**
   * NULL-terminated array of extensions (NULL-terminated strings)
   */
  const char *const *extensions;
  uintptr_t extensions_len;
  /**
   * NULL-terminated array of main fields (NULL-terminated strings)
   */
  const char *const *main_fields;
  uintptr_t main_fields_len;
  /**
   * NULL-terminated array of exports fields (NULL-terminated strings)
   */
  const char *const *exports_fields;
  uintptr_t exports_fields_len;
  /**
   * Path to tsconfig.json (NULL-terminated string, NULL if not provided)
   */
  const char *tsconfig_path;
  /**
   * Whether to enforce extensions
   */
  bool enforce_extension;
  /**
   * Whether request is fully specified
   */
  bool fully_specified;
  /**
   * Whether to resolve symlinks
   */
  bool symlinks;
  /**
   * Prefer relative resolution
   */
  bool prefer_relative;
  /**
   * If true, only resolve to declaration files
   */
  bool declaration_only;
  /**
   * Reserved for future use
   */
  uint64_t _reserved[8];
} CResolveOptions;

/**
 * Resolution result returned to caller
 */
typedef struct CResolution {
  /**
   * Absolute path to resolved file (must be freed by caller with oxc_string_free)
   */
  char *path;
  /**
   * Query string if present (must be freed by caller with oxc_string_free)
   */
  char *query;
  /**
   * Fragment if present (must be freed by caller with oxc_string_free)
   */
  char *fragment;
  /**
   * Error message if error_code != 0 (must be freed by caller with oxc_string_free)
   */
  char *error_message;
  /**
   * Error code (0 = success, non-zero = error)
   */
  int error_code;
  /**
   * Whether resolution used explicit TypeScript extension
   */
  bool resolved_using_ts_extension;
  /**
   * Reserved for future use
   */
  uint64_t _reserved[4];
} CResolution;

#ifdef __cplusplus
extern "C" {
#endif // __cplusplus

/**
 * Create a new resolver instance
 *
 * # Safety
 * - `options` must be a valid pointer to CResolveOptions
 * - All string pointers in options must be valid NULL-terminated UTF-8
 * - Caller must call oxc_resolver_free when done
 */
struct CResolver *oxc_resolver_new(const struct CResolveOptions *options);

/**
 * Resolve a module specifier
 *
 * # Safety
 * - `resolver` must be a valid pointer returned from oxc_resolver_new
 * - `path` must be a valid NULL-terminated UTF-8 string (absolute directory path)
 * - `specifier` must be a valid NULL-terminated UTF-8 string
 * - Caller must free strings in CResolution using oxc_resolution_free
 */
struct CResolution oxc_resolver_resolve(struct CResolver *resolver,
                                        const char *path,
                                        const char *specifier);

/**
 * Resolve a type reference directive
 *
 * # Safety
 * - `resolver` must be a valid pointer returned from oxc_resolver_new
 * - `containing_file` must be a valid NULL-terminated UTF-8 string (absolute file path)
 * - `type_reference` must be a valid NULL-terminated UTF-8 string
 * - Caller must free strings in CResolution using oxc_resolution_free
 */
struct CResolution oxc_resolver_resolve_type_reference_directive(struct CResolver *resolver,
                                                                 const char *containing_file,
                                                                 const char *type_reference);

/**
 * Free a resolver instance
 *
 * # Safety
 * - `resolver` must be a valid pointer returned from oxc_resolver_new
 * - `resolver` must not be used after this call
 */
void oxc_resolver_free(struct CResolver *resolver);

/**
 * Free a C string returned by the resolver
 *
 * # Safety
 * - `s` must be a valid pointer returned from oxc_resolver_resolve
 * - `s` must not be used after this call
 */
void oxc_string_free(char *s);

/**
 * Free all strings in a CResolution
 *
 * # Safety
 * - `resolution` must be a valid pointer to CResolution
 * - Strings in resolution must not be used after this call
 */
void oxc_resolution_free(struct CResolution *resolution);

#ifdef __cplusplus
}  // extern "C"
#endif  // __cplusplus

#endif  /* OXC_RESOLVER_H */
