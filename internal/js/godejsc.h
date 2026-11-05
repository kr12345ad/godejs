/* Minimal JavaScriptCore C API surface used by GodeJS.
 *
 * The system ships libjavascriptcoregtk without development headers, so the
 * declarations we need are spelled out here. Every pointer is declared
 * uintptr_t, not void*: JavaScriptCore NaN-boxes JSValueRef, so `undefined`
 * can be 0xa. An integer ABI keeps such handles out of Go's pointer maps.
 * On x86-64 the integer and pointer calling conventions are identical.
 */
#ifndef GODEJSC_H
#define GODEJSC_H

#include <stddef.h>
#include <stdint.h>

extern uintptr_t JSGlobalContextCreate(uintptr_t globalObjectClass);
extern void JSGlobalContextRelease(uintptr_t ctx);
extern uintptr_t JSContextGetGlobalObject(uintptr_t ctx);

extern uintptr_t JSStringCreateWithUTF8CString(uintptr_t string);
extern uintptr_t JSStringCreateWithCharacters(uintptr_t chars, size_t numChars);
extern void JSStringRelease(uintptr_t string);
extern size_t JSStringGetLength(uintptr_t string);
extern uintptr_t JSStringGetCharactersPtr(uintptr_t string);

extern uintptr_t JSEvaluateScript(uintptr_t ctx, uintptr_t script, uintptr_t thisObject,
                                  uintptr_t sourceURL, int startingLineNumber, uintptr_t exception);
extern uint8_t JSCheckScriptSyntax(uintptr_t ctx, uintptr_t script, uintptr_t sourceURL,
                                   int startingLineNumber, uintptr_t exception);
extern uintptr_t JSValueMakeUndefined(uintptr_t ctx);
extern uintptr_t JSValueMakeNull(uintptr_t ctx);
extern uintptr_t JSValueMakeBoolean(uintptr_t ctx, uint8_t boolean);
extern uintptr_t JSValueMakeNumber(uintptr_t ctx, double number);
extern uintptr_t JSValueMakeString(uintptr_t ctx, uintptr_t string);
extern uintptr_t JSValueMakeFromJSONString(uintptr_t ctx, uintptr_t string);
extern uintptr_t JSValueCreateJSONString(uintptr_t ctx, uintptr_t value, unsigned indent, uintptr_t exception);
extern uint8_t JSValueToBoolean(uintptr_t ctx, uintptr_t value);
extern double JSValueToNumber(uintptr_t ctx, uintptr_t value, uintptr_t exception);
extern uintptr_t JSValueToStringCopy(uintptr_t ctx, uintptr_t value, uintptr_t exception);
extern uintptr_t JSValueToObject(uintptr_t ctx, uintptr_t value, uintptr_t exception);
extern uint8_t JSValueIsStrictEqual(uintptr_t ctx, uintptr_t a, uintptr_t b);
extern uint8_t JSValueIsUndefined(uintptr_t ctx, uintptr_t value);
extern uint8_t JSValueIsNull(uintptr_t ctx, uintptr_t value);
extern uint8_t JSValueIsBoolean(uintptr_t ctx, uintptr_t value);
extern uint8_t JSValueIsNumber(uintptr_t ctx, uintptr_t value);
extern uint8_t JSValueIsString(uintptr_t ctx, uintptr_t value);
extern uint8_t JSValueIsObject(uintptr_t ctx, uintptr_t value);
extern uint8_t JSValueIsArray(uintptr_t ctx, uintptr_t value);
extern uint8_t JSValueIsDate(uintptr_t ctx, uintptr_t value);
extern int JSValueGetType(uintptr_t ctx, uintptr_t value);
extern void JSValueProtect(uintptr_t ctx, uintptr_t value);
extern void JSValueUnprotect(uintptr_t ctx, uintptr_t value);

extern uintptr_t JSObjectMake(uintptr_t ctx, uintptr_t jsClass, uintptr_t data);
extern uintptr_t JSObjectMakeArray(uintptr_t ctx, size_t argumentCount, uintptr_t arguments, uintptr_t exception);
extern uintptr_t JSObjectMakeError(uintptr_t ctx, size_t argumentCount, uintptr_t arguments, uintptr_t exception);
extern uintptr_t JSObjectMakeTypedArray(uintptr_t ctx, int arrayType, size_t length, uintptr_t exception);
extern uintptr_t JSObjectGetTypedArrayBytesPtr(uintptr_t ctx, uintptr_t object, uintptr_t exception);
extern size_t JSObjectGetTypedArrayLength(uintptr_t ctx, uintptr_t object, uintptr_t exception);
extern size_t JSObjectGetArrayBufferByteLength(uintptr_t ctx, uintptr_t object, uintptr_t exception);
extern uintptr_t JSObjectGetArrayBufferBytesPtr(uintptr_t ctx, uintptr_t object, uintptr_t exception);
extern int JSValueGetTypedArrayType(uintptr_t ctx, uintptr_t value, uintptr_t exception);

extern uintptr_t JSObjectGetProperty(uintptr_t ctx, uintptr_t object, uintptr_t propertyName, uintptr_t exception);
extern void JSObjectSetProperty(uintptr_t ctx, uintptr_t object, uintptr_t propertyName, uintptr_t value, int attributes, uintptr_t exception);
extern uint8_t JSObjectHasProperty(uintptr_t ctx, uintptr_t object, uintptr_t propertyName);
extern uint8_t JSObjectDeleteProperty(uintptr_t ctx, uintptr_t object, uintptr_t propertyName, uintptr_t exception);
extern uintptr_t JSObjectGetPropertyAtIndex(uintptr_t ctx, uintptr_t object, unsigned propertyIndex, uintptr_t exception);
extern void JSObjectSetPropertyAtIndex(uintptr_t ctx, uintptr_t object, unsigned propertyIndex, uintptr_t value, uintptr_t exception);
extern uintptr_t JSObjectCallAsFunction(uintptr_t ctx, uintptr_t object, uintptr_t thisObject,
                                        size_t argumentCount, uintptr_t arguments, uintptr_t exception);
extern uintptr_t JSObjectCallAsConstructor(uintptr_t ctx, uintptr_t object,
                                           size_t argumentCount, uintptr_t arguments, uintptr_t exception);
extern uint8_t JSObjectIsFunction(uintptr_t ctx, uintptr_t object);
extern uintptr_t JSObjectGetPrototype(uintptr_t ctx, uintptr_t object);
extern uint8_t JSObjectSetPrototype(uintptr_t ctx, uintptr_t object, uintptr_t value);
extern uintptr_t JSObjectCopyPropertyNames(uintptr_t ctx, uintptr_t object);
extern size_t JSPropertyNameArrayGetCount(uintptr_t array);
extern uintptr_t JSPropertyNameArrayGetNameAtIndex(uintptr_t array, size_t index);
extern void JSPropertyNameArrayRelease(uintptr_t array);

extern uintptr_t godeCallback(uintptr_t ctx, uintptr_t function, uintptr_t thisObject,
                              size_t argumentCount, uintptr_t arguments, uintptr_t exception);

extern uintptr_t JSObjectMakeFunctionWithCallback(uintptr_t ctx, uintptr_t name,
                                                  uintptr_t callAsFunction, uintptr_t exception);
/* constant for the JSObjectCallAsFunctionCallback cast below */
#define GODE_JS_CALLBACK (uintptr_t)godeCallback

#endif
