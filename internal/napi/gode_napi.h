#ifndef GODE_NAPI_H
#define GODE_NAPI_H

#include <dlfcn.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

typedef uintptr_t (*gode_napi_register_fn)(uintptr_t env, uintptr_t exports);
typedef uintptr_t (*gode_napi_callback_fn)(uintptr_t env, uintptr_t info);
typedef void (*gode_napi_execute_fn)(uintptr_t env, uintptr_t data);
typedef void (*gode_napi_complete_fn)(uintptr_t env, int status, uintptr_t data);
typedef void (*gode_napi_finalize_fn)(uintptr_t env, uintptr_t data, uintptr_t hint);
typedef void (*gode_napi_cleanup_fn)(uintptr_t arg);
typedef void (*gode_napi_tsfn_fn)(uintptr_t env, uintptr_t js_callback, uintptr_t context, uintptr_t data);

typedef struct gode_napi_module {
  int nm_version;
  unsigned int nm_flags;
  const char* nm_filename;
  uintptr_t nm_register_func;
  const char* nm_modname;
  void* nm_priv;
  void* reserved[4];
} gode_napi_module;

typedef struct gode_napi_error_info {
  const char* error_message;
  void* engine_reserved;
  uint32_t engine_error_code;
  int error_code;
} gode_napi_error_info;

typedef struct gode_napi_node_version {
  uint32_t major;
  uint32_t minor;
  uint32_t patch;
  const char* release;
} gode_napi_node_version;

typedef struct gode_napi_type_tag {
  uint64_t lower;
  uint64_t upper;
} gode_napi_type_tag;

typedef struct gode_napi_property_descriptor {
  const char* utf8name;
  uintptr_t name;
  uintptr_t method;
  uintptr_t getter;
  uintptr_t setter;
  uintptr_t value;
  int32_t attributes;
  uintptr_t data;
} gode_napi_property_descriptor;

static gode_napi_error_info gode_napi_error;
static gode_napi_node_version gode_napi_version = {24, 22, 2, "node"};

static void* gode_dlopen(const char* path) {
  return dlopen(path, RTLD_NOW | RTLD_LOCAL);
}

static const char* gode_dlerror(void) {
  const char* e = dlerror();
  return e ? e : "dlopen failed";
}

static uintptr_t gode_dlsym(void* handle, const char* name) {
  return (uintptr_t)dlsym(handle, name);
}

static uintptr_t gode_call_register(void* fn, uintptr_t env, uintptr_t exports) {
  return ((gode_napi_register_fn)fn)(env, exports);
}

static uintptr_t gode_call_callback(uintptr_t fn, uintptr_t env, uintptr_t info) {
  return ((gode_napi_callback_fn)fn)(env, info);
}

static void gode_call_execute(uintptr_t fn, uintptr_t env, uintptr_t data) {
  ((gode_napi_execute_fn)fn)(env, data);
}

static void gode_call_complete(uintptr_t fn, uintptr_t env, int status, uintptr_t data) {
  ((gode_napi_complete_fn)fn)(env, status, data);
}

static void gode_call_finalize(uintptr_t fn, uintptr_t env, uintptr_t data, uintptr_t hint) {
  ((gode_napi_finalize_fn)fn)(env, data, hint);
}

static void gode_call_cleanup(uintptr_t fn, uintptr_t arg) {
  ((gode_napi_cleanup_fn)fn)(arg);
}

static void gode_call_tsfn(uintptr_t fn, uintptr_t env, uintptr_t js_callback, uintptr_t context, uintptr_t data) {
  ((gode_napi_tsfn_fn)fn)(env, js_callback, context, data);
}

static uintptr_t gode_module_register_func(const gode_napi_module* mod) {
  return mod ? mod->nm_register_func : 0;
}

static const gode_napi_error_info* gode_last_error_info(void) {
  return &gode_napi_error;
}

static const gode_napi_node_version* gode_node_version(void) {
  return &gode_napi_version;
}

#endif
