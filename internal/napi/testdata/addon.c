// Real Node-API addon used to exercise GodeJS's Go ABI. It is compiled with
// the reference Node 24 headers and has no GodeJS-specific code: every napi_*
// symbol is resolved from the gode executable at dlopen time.
#include <node_api.h>

#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define AUTO NAPI_AUTO_LENGTH

static napi_value Add(napi_env env, napi_callback_info info) {
  size_t argc = 2;
  napi_value argv[2];
  napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
  double a = 0, b = 0;
  napi_get_value_double(env, argv[0], &a);
  napi_get_value_double(env, argv[1], &b);
  napi_value out;
  napi_create_double(env, a + b, &out);
  return out;
}

static napi_value Greet(napi_env env, napi_callback_info info) {
  size_t argc = 1;
  napi_value argv[1];
  napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
  char name[128] = {0};
  size_t len = 0;
  napi_get_value_string_utf8(env, argv[0], name, sizeof(name), &len);
  char buf[256];
  snprintf(buf, sizeof(buf), "hello %s", name);
  napi_value out;
  napi_create_string_utf8(env, buf, AUTO, &out);
  return out;
}

static napi_value GetComputed(napi_env env, napi_callback_info info) {
  (void)info;
  napi_value out;
  napi_create_int32(env, 7, &out);
  return out;
}

static napi_value Info(napi_env env, napi_callback_info info) {
  (void)info;
  napi_value obj, s, n;
  napi_create_object(env, &obj);
  napi_create_string_utf8(env, "gode", AUTO, &s);
  napi_set_named_property(env, obj, "name", s);
  napi_create_int32(env, 42, &n);
  napi_set_named_property(env, obj, "n", n);
  napi_property_descriptor props[] = {
      {"computed", NULL, NULL, GetComputed, NULL, NULL, napi_enumerable, NULL},
  };
  napi_define_properties(env, obj, 1, props);
  return obj;
}

static void FinalizeState(napi_env env, void* data, void* hint) {
  (void)env;
  (void)hint;
  printf("[napi] finalized counter state %d\n", *(int*)data);
  free(data);
}

static void FinalizeInt(napi_env env, void* data, void* hint) {
  (void)env;
  (void)hint;
  free(data);
}

static napi_value CounterCtor(napi_env env, napi_callback_info info) {
  size_t argc = 1;
  napi_value argv[1];
  napi_value self;
  napi_get_cb_info(env, info, &argc, argv, &self, NULL);
  napi_value new_target;
  napi_get_new_target(env, info, &new_target);
  if (new_target == NULL) {
    napi_throw_type_error(env, NULL, "Counter must be called with new");
    return NULL;
  }
  int32_t start = 0;
  if (argc > 0) {
    napi_get_value_int32(env, argv[0], &start);
  }
  napi_value v, is_new;
  napi_create_int32(env, start, &v);
  napi_set_named_property(env, self, "value", v);
  napi_get_boolean(env, new_target != NULL, &is_new);
  napi_set_named_property(env, self, "isNew", is_new);
  int* state = (int*)malloc(sizeof(int));
  *state = start;
  napi_wrap(env, self, state, FinalizeState, NULL, NULL);
  return self;
}

static napi_value CounterInc(napi_env env, napi_callback_info info) {
  napi_value self;
  napi_get_cb_info(env, info, NULL, NULL, &self, NULL);
  napi_value v;
  napi_get_named_property(env, self, "value", &v);
  int32_t cur = 0;
  napi_get_value_int32(env, v, &cur);
  napi_value out;
  napi_create_int32(env, cur + 1, &out);
  napi_set_named_property(env, self, "value", out);
  return out;
}

static napi_value CounterValueStr(napi_env env, napi_callback_info info) {
  napi_value self;
  napi_get_cb_info(env, info, NULL, NULL, &self, NULL);
  napi_value v;
  napi_get_named_property(env, self, "value", &v);	napi_value text;
	napi_coerce_to_string(env, v, &text);
	char buf[32] = {0};
	size_t len = 0;
	napi_get_value_string_utf8(env, text, buf, sizeof(buf), &len);
  char out[64];
  snprintf(out, sizeof(out), "value=%s", buf);
  napi_value res;
  napi_create_string_utf8(env, out, AUTO, &res);
  return res;
}

static napi_value Bump(napi_env env, napi_callback_info info) {
  size_t argc = 1;
  napi_value argv[1];
  napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
  void* data = NULL;
  size_t len = 0;
  napi_get_buffer_info(env, argv[0], &data, &len);
  unsigned char* p = (unsigned char*)data;
  for (size_t i = 0; i < len; i++) {
    p[i] += 1;
  }
  napi_value out;
  napi_create_uint32(env, (uint32_t)len, &out);
  return out;
}

static napi_value SumF64(napi_env env, napi_callback_info info) {
  size_t argc = 1;
  napi_value argv[1];
  napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
  napi_typedarray_type type;
  size_t length = 0, offset = 0;
  void* data = NULL;
  napi_value buffer;
  napi_get_typedarray_info(env, argv[0], &type, &length, &data, &buffer, &offset);
  double* p = (double*)data;
  double total = 0;
  if (type == napi_float64_array) {
    for (size_t i = 0; i < length; i++) {
      total += p[i];
    }
  }
  napi_value out;
  napi_create_double(env, total, &out);
  return out;
}

static napi_value BigAdd(napi_env env, napi_callback_info info) {
  size_t argc = 2;
  napi_value argv[2];
  napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
  int64_t a = 0, b = 0;
  bool lossless = false;
  napi_get_value_bigint_int64(env, argv[0], &a, &lossless);
  napi_get_value_bigint_int64(env, argv[1], &b, &lossless);
  napi_value out;
  napi_create_bigint_int64(env, a + b, &out);
  return out;
}

static napi_value DateMs(napi_env env, napi_callback_info info) {
  size_t argc = 1;
  napi_value argv[1];
  napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
  double ms = 0;
  napi_get_value_double(env, argv[0], &ms);
  napi_value d;
  napi_create_date(env, ms, &d);
  bool is_date = false;
  napi_is_date(env, d, &is_date);
  double back = -1;
  napi_get_date_value(env, d, &back);
  napi_value out;
  napi_create_double(env, is_date ? back : -1, &out);
  return out;
}

static napi_value PromiseValue(napi_env env, napi_callback_info info) {
  (void)info;
  napi_value promise;
  napi_deferred deferred;
  napi_create_promise(env, &deferred, &promise);
  napi_value v;
  napi_create_string_utf8(env, "resolved", AUTO, &v);
  napi_resolve_deferred(env, deferred, v);
  return promise;
}

typedef struct {
  napi_async_work work;
  napi_deferred deferred;
  int value;
} work_ctx;

static void AsyncExecute(napi_env env, void* data) {
  (void)env;
  work_ctx* ctx = (work_ctx*)data;
  ctx->value *= 2;
}

static void AsyncComplete(napi_env env, napi_status status, void* data) {
  work_ctx* ctx = (work_ctx*)data;
  napi_value result;
  if (status == napi_ok) {
    napi_create_int32(env, ctx->value, &result);
    napi_resolve_deferred(env, ctx->deferred, result);
  } else {
    napi_value msg;
    napi_create_string_utf8(env, "cancelled", AUTO, &msg);
    napi_reject_deferred(env, ctx->deferred, msg);
  }
  napi_delete_async_work(env, ctx->work);
  free(ctx);
}

static napi_value AsyncDouble(napi_env env, napi_callback_info info) {
  size_t argc = 1;
  napi_value argv[1];
  napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
  int32_t v = 0;
  napi_get_value_int32(env, argv[0], &v);
  work_ctx* ctx = (work_ctx*)calloc(1, sizeof(work_ctx));
  ctx->value = v;
  napi_value promise, name;
  napi_create_promise(env, &ctx->deferred, &promise);
  napi_create_string_utf8(env, "asyncDouble", AUTO, &name);
  napi_create_async_work(env, NULL, name, AsyncExecute, AsyncComplete, ctx, &ctx->work);
  napi_queue_async_work(env, ctx->work);
  return promise;
}

static napi_value WrapMe(napi_env env, napi_callback_info info) {
  (void)info;
  napi_value obj;
  napi_create_object(env, &obj);
  int* p = (int*)malloc(sizeof(int));
  *p = 99;
  napi_wrap(env, obj, p, FinalizeInt, NULL, NULL);
  return obj;
}

static napi_value UnwrapMe(napi_env env, napi_callback_info info) {
  size_t argc = 1;
  napi_value argv[1];
  napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
  void* p = NULL;
  if (napi_unwrap(env, argv[0], &p) != napi_ok || p == NULL) {
    napi_throw_error(env, NULL, "unwrap failed");
    return NULL;
  }
  napi_value out;
  napi_create_int32(env, *(int*)p, &out);
  return out;
}

static napi_value Boom(napi_env env, napi_callback_info info) {
  (void)info;
  napi_throw_error(env, "EBOOM", "boom!");
  return NULL;
}

static napi_threadsafe_function g_tsfn = NULL;

static void CallJS(napi_env env, napi_value cb, void* context, void* data) {
  (void)context;
  napi_value global, arg, undef;
  napi_get_global(env, &global);
  napi_get_undefined(env, &undef);
  napi_create_uint32(env, (uint32_t)(uintptr_t)data, &arg);
  napi_call_function(env, global, cb, 1, &arg, &undef);
}

static napi_value TsfnEmit(napi_env env, napi_callback_info info) {
  size_t argc = 2;
  napi_value argv[2];
  napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
  uint32_t v = 0;
  napi_get_value_uint32(env, argv[1], &v);
  napi_value name;
  napi_create_string_utf8(env, "gode-test", AUTO, &name);
  napi_status s = napi_create_threadsafe_function(
      env, argv[0], NULL, name, 0, 1, NULL, NULL, NULL, CallJS, &g_tsfn);
  if (s != napi_ok) {
    napi_throw_error(env, NULL, "tsfn create failed");
    return NULL;
  }
  s = napi_call_threadsafe_function(g_tsfn, (void*)(uintptr_t)v, napi_tsfn_nonblocking);
  if (s != napi_ok) {
    napi_throw_error(env, NULL, "tsfn call failed");
    return NULL;
  }
  napi_release_threadsafe_function(g_tsfn, napi_tsfn_release);
  g_tsfn = NULL;
  return NULL;
}

static napi_value InstanceCheck(napi_env env, napi_callback_info info) {
  size_t argc = 2;
  napi_value argv[2];
  napi_get_cb_info(env, info, &argc, argv, NULL, NULL);
  bool ok = false;
  napi_instanceof(env, argv[0], argv[1], &ok);
  napi_value out;
  napi_get_boolean(env, ok, &out);
  return out;
}

NAPI_MODULE_INIT() {
  struct {
    const char* name;
    napi_callback cb;
    napi_value* slot;
  } funcs[] = {
      {"add", Add, NULL},
      {"greet", Greet, NULL},
      {"info", Info, NULL},
      {"bump", Bump, NULL},
      {"sumF64", SumF64, NULL},
      {"bigAdd", BigAdd, NULL},
      {"dateMs", DateMs, NULL},
      {"promiseValue", PromiseValue, NULL},
      {"asyncDouble", AsyncDouble, NULL},
      {"wrapMe", WrapMe, NULL},
      {"unwrapMe", UnwrapMe, NULL},
      {"boom", Boom, NULL},
      {"tsfnEmit", TsfnEmit, NULL},
      {"instanceCheck", InstanceCheck, NULL},
  };
  for (size_t i = 0; i < sizeof(funcs) / sizeof(funcs[0]); i++) {
    napi_value fn;
    napi_create_function(env, funcs[i].name, AUTO, funcs[i].cb, NULL, &fn);
    napi_set_named_property(env, exports, funcs[i].name, fn);
  }

  napi_property_descriptor props[] = {
      {"inc", NULL, CounterInc, NULL, NULL, NULL, napi_default_method, NULL},
      {"valueStr", NULL, NULL, CounterValueStr, NULL, NULL,
       napi_default | napi_enumerable, NULL},
  };
  napi_value cls;
  if (napi_define_class(env, "Counter", AUTO, CounterCtor, NULL,
                        sizeof(props) / sizeof(props[0]), props, &cls) == napi_ok) {
    napi_set_named_property(env, exports, "Counter", cls);
  }
  return exports;
}
