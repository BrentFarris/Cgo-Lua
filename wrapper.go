package lua

/*
//#cgo LDFLAGS: -lm -ldl
#ifdef __linux__
#define LUA_USE_LINUX
#endif
#include <stdlib.h>
#include <stdint.h>
#include "lua.h"
#include "lualib.h"
#include "lauxlib.h"
extern int panic_callback(lua_State* L);
extern int pcallk_callback(lua_State* L, int status, lua_KContext ctx);
extern int cclosure_callback(lua_State* L);
extern int print_stack(lua_State* lua);
*/
import "C"
import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"unsafe"
)

const (
	LUA_ERRERR          = 5
	LUA_ERRFILE         = LUA_ERRERR + 1
	LUA_ERRMEM          = 4
	LUA_ERRRUN          = 2
	LUA_ERRSYNTAX       = 3
	LUA_HOOKCALL        = 0
	LUA_HOOKCOUNT       = 3
	LUA_HOOKLINE        = 2
	LUA_HOOKRET         = 1
	LUA_HOOKTAILCALL    = 4
	LUA_LOADED_TABLE    = "_LOADED"
	LUA_MASKCALL        = 1 << LUA_HOOKCALL
	LUA_MASKCOUNT       = 1 << LUA_HOOKCOUNT
	LUA_MASKLINE        = 1 << LUA_HOOKLINE
	LUA_MASKRET         = 1 << LUA_HOOKRET
	LUA_MAXINTEGER      = math.MaxInt64
	LUA_MININTEGER      = math.MinInt64
	LUA_MINSTACK        = 20
	LUA_MULTRET         = -1
	LUA_NOREF           = -2
	LUA_OK              = 0
	LUA_OPADD           = 0
	LUA_OPBAND          = 7
	LUA_OPBNOT          = 13
	LUA_OPBOR           = 8
	LUA_OPBXOR          = 9
	LUA_OPDIV           = 5
	LUA_OPEQ            = 0
	LUA_OPIDIV          = 6
	LUA_OPLE            = 2
	LUA_OPLT            = 1
	LUA_OPMOD           = 3
	LUA_OPMUL           = 2
	LUA_OPPOW           = 4
	LUA_OPSHL           = 10
	LUA_OPSHR           = 11
	LUA_OPSUB           = 1
	LUA_OPUNM           = 12
	LUA_PRELOAD_TABLE   = "_PRELOAD"
	LUA_REFNIL          = -1
	LUAI_MAXSTACK       = 1000000
	LUA_REGISTRYINDEX   = -LUAI_MAXSTACK - 1000
	LUA_RIDX_GLOBALS    = 2
	LUA_RIDX_MAINTHREAD = 1
	LUA_TBOOLEAN        = 1
	LUA_TFUNCTION       = 6
	LUA_TLIGHTUSERDATA  = 2
	LUA_TNIL            = 0
	LUA_TNONE           = -1
	LUA_TNUMBER         = 3
	LUA_TSTRING         = 4
	LUA_TTABLE          = 5
	LUA_TTHREAD         = 8
	LUA_TUSERDATA       = 7
	LUA_YIELD           = 1
	LUA_VERSION_NUM     = 504
	LUAL_NUMSIZES       = 8*16 + 8 //(sizeof(lua_Integer)*16 + sizeof(lua_Number))
	//LUA_USE_APICHECK    = C.LUA_USE_APICHECK
	//LUAL_BUFFERSIZE     = C.LUAL_BUFFERSIZE
)

type luaClosure struct {
	id   int64
	call func(L *LuaState) int
}

type LuaState struct {
	closureId int64
	luaState  *C.lua_State
	onPanic   func() int
	onPCallK  func() int
	closures  map[int64]luaClosure
}

var luaMap = make(map[*C.lua_State]*LuaState)

func NewLuaState() *LuaState {
	L := &LuaState{
		closureId: 0,
		luaState:  C.luaL_newstate(),
		onPanic:   nil,
		onPCallK:  nil,
		closures:  make(map[int64]luaClosure),
	}
	L.OpenLibs()
	luaMap[L.luaState] = L
	C.lua_pushcclosure(L.luaState, (C.lua_CFunction)(C.cclosure_callback), 0)
	L.SetGlobal("_____closure_fn")
	return L
}

func (l *LuaState) Close() {
	C.lua_close(l.luaState)
}

func (l *LuaState) Call(nargs, nresults int) {
	C.lua_callk(l.luaState, C.int(nargs), C.int(nresults), 0, nil)
}

func (l *LuaState) AbsIndex(idx int) int {
	return int(C.lua_absindex(l.luaState, C.int(idx)))
}

func (l *LuaState) Arith(op int) {
	C.lua_arith(l.luaState, C.int(op))
}

//export panic_callback
func panic_callback(l *C.lua_State) C.int {
	L := luaMap[l]
	if L.onPanic != nil {
		return C.int(L.onPanic())
	}
	return 0
}

func (l *LuaState) AtPanic(onPanic func() int) {
	l.onPanic = onPanic
	C.lua_atpanic(l.luaState, (C.lua_CFunction)(C.panic_callback))
}

func (l *LuaState) CheckStack(n int) bool {
	return C.lua_checkstack(l.luaState, C.int(n)) == LUA_OK
}

func (l *LuaState) CloseSlot(idx int) {
	C.lua_closeslot(l.luaState, C.int(idx))
}

func (l *LuaState) Compare(idx1, idx2 int, op int) bool {
	return C.lua_compare(l.luaState, C.int(idx1), C.int(idx2), C.int(op)) == LUA_OK
}

func (l *LuaState) Concat(n int) {
	C.lua_concat(l.luaState, C.int(n))
}

func (l *LuaState) Copy(fromIdx, toIdx int) {
	C.lua_copy(l.luaState, C.int(fromIdx), C.int(toIdx))
}

func (l *LuaState) CreateTable(nArr, nRec int) {
	C.lua_createtable(l.luaState, C.int(nArr), C.int(nRec))
}

func (l *LuaState) Dump(writer func([]byte) int) int {
	// TODO:
	return 0
}

func (l *LuaState) Error() int {
	return int(C.lua_error(l.luaState))
}

func (l *LuaState) GC(what int, data ...int) int {
	panic("not implemented, variadic function, use directly")
}

func (l *LuaState) GetAllocF(ud *unsafe.Pointer) {
	//C.lua_getallocf(l.luaState, ud)
	panic("not implemented")
}

func (l *LuaState) GetExtraSpace() {
	//C.lua_getextraspace(l.luaState)
	panic("not implemented")
}

func (l *LuaState) GetField(idx int, k string) bool {
	cs := C.CString(k)
	defer C.free(unsafe.Pointer(cs))
	return C.lua_getfield(l.luaState, C.int(idx), cs) == LUA_OK
}

func (l *LuaState) GetGlobal(name string) {
	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))
	C.lua_getglobal(l.luaState, cs)
}

func (l *LuaState) Gethook() {
	//C.lua_gethook(l.luaState)
	panic("not implemented")
}

func (l *LuaState) GetHookCount() int {
	return int(C.lua_gethookcount(l.luaState))
}

func (l *LuaState) GetHookMask() int {
	return int(C.lua_gethookmask(l.luaState))
}

func (l *LuaState) GetI(idx int, n int64) int {
	return int(C.lua_geti(l.luaState, C.int(idx), C.lua_Integer(n)))
}

func (l *LuaState) GetInfo() {
	//C.lua_getinfo(l.luaState)
	panic("not implemented, debug stuff")
}

func (l *LuaState) GetIUserValue(idx, n int) int {
	return int(C.lua_getiuservalue(l.luaState, C.int(idx), C.int(n)))
}

func (l *LuaState) GetLocal() {
	//C.lua_getlocal(l.luaState)
	panic("not implemented, debug stuff")
}

func (l *LuaState) GetMetaTable(objindex int) int {
	return int(C.lua_getmetatable(l.luaState, C.int(objindex)))
}

func (l *LuaState) GetStack() {
	//C.lua_getstack(l.luaState)
	panic("not implemented, debug stuff")
}

func (l *LuaState) GetTable(idx int) int {
	return int(C.lua_gettable(l.luaState, C.int(idx)))
}

func (l *LuaState) GetTop() int {
	return int(C.lua_gettop(l.luaState))
}

func (l *LuaState) GetUpValue(funcindex, n int) string {
	return C.GoString(C.lua_getupvalue(l.luaState, C.int(funcindex), C.int(n)))
}

func (l *LuaState) Insert(idx int) {
	l.Rotate(idx, 1)
}

func (l *LuaState) IsBoolean(n int) bool {
	return l.Type(n) == LUA_TBOOLEAN
}

func (l *LuaState) IsCFunction(n int) bool {
	return C.lua_iscfunction(l.luaState, C.int(n)) == LUA_OK
}

func (l *LuaState) IsFunction(n int) bool {
	return l.Type(n) == LUA_TFUNCTION
}

func (l *LuaState) IsInteger(idx int) bool {
	return C.lua_isinteger(l.luaState, C.int(idx)) == LUA_OK
}

func (l *LuaState) IsLightUserData(n int) bool {
	return l.Type(n) == LUA_TLIGHTUSERDATA
}

func (l *LuaState) IsNil(n int) bool {
	return l.Type(n) == LUA_TNIL
}

func (l *LuaState) IsNone(n int) bool {
	return l.Type(n) == LUA_TNONE
}

func (l *LuaState) IsNoneOrNil(n int) bool {
	return l.Type(n) <= 0
}

func (l *LuaState) IsNumber(idx int) bool {
	return C.lua_isnumber(l.luaState, C.int(idx)) == LUA_OK
}

func (l *LuaState) IsString(idx int) bool {
	return C.lua_isstring(l.luaState, C.int(idx)) == LUA_OK
}

func (l *LuaState) IsTable(n int) bool {
	return l.Type(n) == LUA_TTABLE
}

func (l *LuaState) IsThread(n int) bool {
	return l.Type(n) == LUA_TTHREAD
}

func (l *LuaState) IsUserData(idx int) bool {
	return C.lua_isuserdata(l.luaState, C.int(idx)) == LUA_OK
}

func (l *LuaState) IsYieldable() bool {
	return C.lua_isyieldable(l.luaState) == LUA_OK
}

func (l *LuaState) Len(idx int) {
	C.lua_len(l.luaState, C.int(idx))
}

func (l *LuaState) Load() {
	//C.lua_load(l.luaState)
	panic("not implemented")
}

func (l *LuaState) NewState() {
	//C.lua_newstate(l.luaState)
	panic("not implemented")
}

func (l *LuaState) NewTable() {
	l.CreateTable(0, 0)
}

func (l *LuaState) NewThread() LuaState {
	return LuaState{
		luaState: C.lua_newthread(l.luaState),
	}
}

func (l *LuaState) NewUserDataUV(sz int64, nuvalue int) unsafe.Pointer {
	return C.lua_newuserdatauv(l.luaState, C.size_t(sz), C.int(nuvalue))
}

func (l *LuaState) Next(idx int) int {
	return int(C.lua_next(l.luaState, C.int(idx)))
}

func (l *LuaState) NumberToInteger(n int64, p *int) bool {
	if n >= LUA_MININTEGER && n < LUA_MAXINTEGER {
		*p = int(n)
		return true
	}
	return false
}

func (l *LuaState) PCall(n, r, f int) bool {
	return l.PCallK(n, r, f, 0, nil)
}

//export pcallk_callback
func pcallk_callback(l *C.lua_State, status C.int, ctx C.lua_KContext) C.int {
	L := luaMap[l]
	if L.onPCallK != nil {
		return C.int(L.onPCallK())
	}
	return 0
}

func (l *LuaState) PCallK(nargs, nresults, errfunc, ctx int, k func() int) bool {
	l.onPCallK = k
	return C.lua_pcallk(l.luaState, C.int(nargs), C.int(nresults), C.int(errfunc),
		C.lua_KContext(ctx), (C.lua_KFunction)(C.pcallk_callback)) == LUA_OK
}

func (l *LuaState) Pop(n int) {
	l.SetTop(-n - 1)
}

func (l *LuaState) PushBoolean(b bool) {
	if b {
		C.lua_pushboolean(l.luaState, 1)
	} else {
		C.lua_pushboolean(l.luaState, 0)
	}
}

//export cclosure_callback
func cclosure_callback(l *C.lua_State) C.int {
	L := luaMap[l]
	closureId := L.ToInteger(1)
	c, ok := L.closures[closureId]
	if ok {
		L.Remove(1)
		return C.int(c.call(L))
	} else {
		return 0
	}
}

func (l *LuaState) PushCClosure(fn C.lua_CFunction, n int) {
	C.lua_pushcclosure(l.luaState, fn, C.int(n))
}

func (l *LuaState) PushCFunction(fn C.lua_CFunction) {
	l.PushCClosure(fn, 0)
}

func (l *LuaState) PushFunction(fn func(L *LuaState) int) {
	l.closures[l.closureId] = luaClosure{
		id:   l.closureId,
		call: fn,
	}
	l.DoString(fmt.Sprintf("return function(...) return _____closure_fn(%d, ...) end", l.closureId))
	l.closureId++
}

func (l *LuaState) PushFString(fmt string, a ...any) {
	//C.lua_pushfstring(l.luaState)
	panic("not implemented")
}

func (l *LuaState) PushGlobalTable() {
	l.RawGetI(LUA_REGISTRYINDEX, LUA_RIDX_GLOBALS)
}

func (l *LuaState) PushInteger(n int64) {
	C.lua_pushinteger(l.luaState, C.lua_Integer(n))
}

func (l *LuaState) PushInt(n int) {
	l.PushInteger(int64(n))
}

func (l *LuaState) PushLightUserData(ptr unsafe.Pointer) {
	C.lua_pushlightuserdata(l.luaState, unsafe.Pointer(uintptr(ptr)))
}

func (l *LuaState) PushUserDataAddress(ptr unsafe.Pointer) {
	l.PushInteger(int64(uintptr(ptr)))
}

func (l *LuaState) PushLiteral(s string) {
	l.PushString(s)
}

func (l *LuaState) PushLString(s string) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	C.lua_pushlstring(l.luaState, cs, C.size_t(len(s)))
}

func (l *LuaState) PushNil() {
	C.lua_pushnil(l.luaState)
}

func (l *LuaState) PushNumber(n float64) {
	C.lua_pushnumber(l.luaState, C.double(n))
}

func (l *LuaState) PushFloat32(n float32) {
	C.lua_pushnumber(l.luaState, C.double(n))
}

func (l *LuaState) PushString(s string) string {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return C.GoString(C.lua_pushstring(l.luaState, cs))
}

func (l *LuaState) PushThread() int {
	return int(C.lua_pushthread(l.luaState))
}

func (l *LuaState) PushValue(idx int) {
	C.lua_pushvalue(l.luaState, C.int(idx))
}

func (l *LuaState) PushVFString(fmt string, a ...any) string {
	//C.lua_pushvfstring(l.luaState)
	panic("not implemented")
}

func (l *LuaState) RawEqual(idx1, idx2 int) int {
	return int(C.lua_rawequal(l.luaState, C.int(idx1), C.int(idx2)))
}

func (l *LuaState) RawGet(idx int) int {
	return int(C.lua_rawget(l.luaState, C.int(idx)))
}

func (l *LuaState) RawGetI(idx int, n int64) int {
	return int(C.lua_rawgeti(l.luaState, C.int(idx), C.lua_Integer(n)))
}

func (l *LuaState) RawGetP(idx int, p unsafe.Pointer) int {
	return int(C.lua_rawgetp(l.luaState, C.int(idx), p))
}

func (l *LuaState) RawLen(idx int) uint64 {
	return uint64(C.lua_rawlen(l.luaState, C.int(idx)))
}

func (l *LuaState) RawSet(idx int) {
	C.lua_rawset(l.luaState, C.int(idx))
}

func (l *LuaState) RawSetI(idx int, n int64) {
	C.lua_rawseti(l.luaState, C.int(idx), C.lua_Integer(n))
}

func (l *LuaState) RawSetP(idx int, p unsafe.Pointer) {
	C.lua_rawsetp(l.luaState, C.int(idx), p)
}

func (l *LuaState) Register(n string, f func(L *LuaState) int) {
	l.PushFunction(f)
	l.SetGlobal(n)
}

func (l *LuaState) Remove(idx int) {
	l.Rotate(idx, -1)
	l.Pop(1)
}

func (l *LuaState) Replace(idx int) {
	l.Copy(-1, idx)
	l.Pop(1)
}

func (l *LuaState) ResetThread() int {
	return int(C.lua_resetthread(l.luaState))
}

func (l *LuaState) Resume(from LuaState, narg int, nres *int) int {
	// TODO:  Review this
	ci := C.int(0)
	res := int(C.lua_resume(l.luaState, from.luaState, C.int(narg), &ci))
	*nres = int(ci)
	return res
}

func (l *LuaState) Rotate(idx, n int) {
	C.lua_rotate(l.luaState, C.int(idx), C.int(n))
}

func (l *LuaState) SetAllocF() {
	//C.lua_setallocf(l.luaState)
	panic("not implemented")
}

func (l *LuaState) SetField(idx int, k string) {
	cs := C.CString(k)
	defer C.free(unsafe.Pointer(cs))
	C.lua_setfield(l.luaState, C.int(idx), cs)
}

func (l *LuaState) SetGlobal(name string) {
	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))
	C.lua_setglobal(l.luaState, cs)
}

func (l *LuaState) SetHook() {
	//C.lua_sethook(l.luaState)
	panic("not implemented")
}

func (l *LuaState) SetI(idx int, n int64) {
	C.lua_seti(l.luaState, C.int(idx), C.lua_Integer(n))
}

func (l *LuaState) SetIUserValue(idx, n int) int {
	return int(C.lua_setiuservalue(l.luaState, C.int(idx), C.int(n)))
}

func (l *LuaState) SetLocal() string {
	//C.lua_setlocal(l.luaState)
	panic("not implemented")
}

func (l *LuaState) SetMetaTable(objindex int) int {
	return int(C.lua_setmetatable(l.luaState, C.int(objindex)))
}

func (l *LuaState) SetTable(idx int) {
	C.lua_settable(l.luaState, C.int(idx))
}

func (l *LuaState) SetTop(idx int) {
	C.lua_settop(l.luaState, C.int(idx))
}

func (l *LuaState) SetUpValue(funcindex, n int) string {
	return C.GoString(C.lua_setupvalue(l.luaState, C.int(funcindex), C.int(n)))
}

func (l *LuaState) SetWarnF() {
	//C.lua_setwarnf(l.luaState)
	panic("not implemented")
}

func (l *LuaState) Status() int {
	return int(C.lua_status(l.luaState))
}

func (l *LuaState) StringToNumber(s string) uint64 {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return uint64(C.lua_stringtonumber(l.luaState, cs))
}

func (l *LuaState) ToBoolean(idx int) bool {
	return C.lua_toboolean(l.luaState, C.int(idx)) == LUA_OK
}

func (l *LuaState) ToCFunction(idx int) {
	//C.lua_tocfunction(l.luaState)
	panic("not implemented")
}

func (l *LuaState) ToClose(idx int) {
	C.lua_toclose(l.luaState, C.int(idx))
}

func (l *LuaState) ToInteger(i int) int64 {
	res, _ := l.ToIntegerX(i)
	return res
}

func (l *LuaState) ToIntegerX(idx int) (int64, bool) {
	// TODO:  Review this
	ci := C.int(0)
	res := int64(C.lua_tointegerx(l.luaState, C.int(idx), &ci))
	return res, int(ci) != 0
}

func (l *LuaState) ToLString(idx int) string {
	// TODO:  Review this
	len := C.size_t(0)
	res := C.lua_tolstring(l.luaState, C.int(idx), &len)
	if res == nil {
		return ""
	} else {
		return C.GoString(res)
	}
}

func (l *LuaState) ToNumber(idx int) float64 {
	res, _ := l.ToNumberX(idx)
	return res
}

func (l *LuaState) ToNumberX(idx int) (float64, bool) {
	ci := C.int(0)
	res := C.lua_tonumberx(l.luaState, C.int(idx), &ci)
	return float64(res), int(ci) != 0
}

func (l *LuaState) ToPointer(idx int) unsafe.Pointer {
	return unsafe.Pointer(C.lua_topointer(l.luaState, C.int(idx)))
}

func (l *LuaState) ToString(idx int) string {
	return l.ToLString(idx)
}

func (l *LuaState) ToThread(idx int) LuaState {
	return LuaState{
		luaState: C.lua_tothread(l.luaState, C.int(idx)),
	}
}

func (l *LuaState) ToUserData(idx int) unsafe.Pointer {
	return C.lua_touserdata(l.luaState, C.int(idx))
}

func (l *LuaState) ToUserDataAddress(idx int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(l.ToInteger(idx)))
}

func (l *LuaState) Type(n int) int {
	return int(C.lua_type(l.luaState, C.int(n)))
}

func (l *LuaState) TypeName(tp int) string {
	return C.GoString(C.lua_typename(l.luaState, C.int(tp)))
}

func (l *LuaState) UpValueId(fidx, n int) unsafe.Pointer {
	return C.lua_upvalueid(l.luaState, C.int(fidx), C.int(n))
}

func (l *LuaState) UpValueIndex(idx int) int {
	return LUA_REGISTRYINDEX - idx
}

func (l *LuaState) UpValueJoin(fidx1, n1, fidx2, n2 int) {
	C.lua_upvaluejoin(l.luaState, C.int(fidx1), C.int(n1), C.int(fidx2), C.int(n2))
}

func (l *LuaState) Version() int {
	return int(C.lua_version(l.luaState))
}

func (l *LuaState) Warning(msg string, tocont int) {
	cs := C.CString(msg)
	defer C.free(unsafe.Pointer(cs))
	C.lua_warning(l.luaState, cs, C.int(tocont))
}

func (l *LuaState) XMove(to LuaState, n int) {
	C.lua_xmove(l.luaState, to.luaState, C.int(n))
}

func (l *LuaState) Yield(nresults int) {
	l.YieldK(nresults, 0, nil)
}

func (l *LuaState) YieldK(nresults int, ctx int, k func()) int {
	//C.lua_yieldk(l.luaState)
	panic("not implemented")
}

func (l *LuaState) AddChar() {
	//luaL_addchar(l.luaState)
	panic("not implemented")
}

func (l *LuaState) AddGSub() {
	//luaL_addgsub(l.luaState)
	panic("not implemented")
}

func (l *LuaState) AddLString() {
	//luaL_addlstring(l.luaState)
	panic("not implemented")
}

func (l *LuaState) AddSize() {
	//luaL_addsize(l.luaState)
	panic("not implemented")
}

func (l *LuaState) AddString() {
	//luaL_addstring(l.luaState)
	panic("not implemented")
}

func (l *LuaState) AddValue() {
	//luaL_addvalue(l.luaState)
	panic("not implemented")
}

func (l *LuaState) ArgCheck() {
	//luaL_argcheck(l.luaState)
	panic("not implemented")
}

func (l *LuaState) ArgError() {
	//luaL_argerror(l.luaState)
	panic("not implemented")
}

func (l *LuaState) ArgExpected() {
	//luaL_argexpected(l.luaState)
	panic("not implemented")
}

func (l *LuaState) BuffAddr() {
	//luaL_buffaddr(l.luaState)
	panic("not implemented")
}

func (l *LuaState) BuffInit() {
	//luaL_buffinit(l.luaState)
	panic("not implemented")
}

func (l *LuaState) BuffInitSize() {
	//luaL_buffinitsize(l.luaState)
	panic("not implemented")
}

func (l *LuaState) BuffLen() {
	//luaL_bufflen(l.luaState)
	panic("not implemented")
}

func (l *LuaState) BuffSub() {
	//luaL_buffsub(l.luaState)
	panic("not implemented")
}

func (l *LuaState) CallMeta(obj int, e string) bool {
	cs := C.CString(e)
	defer C.free(unsafe.Pointer(cs))
	return int(C.luaL_callmeta(l.luaState, C.int(obj), cs)) == LUA_OK
}

func (l *LuaState) CheckAny(arg int) {
	C.luaL_checkany(l.luaState, C.int(arg))
}

func (l *LuaState) CheckInteger(arg int) int64 {
	return int64(C.luaL_checkinteger(l.luaState, C.int(arg)))
}

func (l *LuaState) CheckLString(arg int) string {
	//len := C.size_t(0)
	//C.luaL_checklstring(l.luaState, C.int(arg), &len)

	// TODO:  Make sure this is always null terminated
	return C.GoString(C.luaL_checklstring(l.luaState, C.int(arg), nil))
}

func (l *LuaState) CheckNumber(arg int) float64 {
	return float64(C.luaL_checknumber(l.luaState, C.int(arg)))
}

func (l *LuaState) CheckOption() {
	//luaL_checkoption(l.luaState)
	panic("not implemented")
}

func (l *LuaState) LCheckStack(sz int, msg string) {
	cs := C.CString(msg)
	defer C.free(unsafe.Pointer(cs))
	C.luaL_checkstack(l.luaState, C.int(sz), cs)
}

func (l *LuaState) CheckString(arg int) string {
	return l.CheckLString(arg)
}

func (l *LuaState) CheckType(arg, t int) {
	C.luaL_checktype(l.luaState, C.int(arg), C.int(t))
}

func (l *LuaState) CheckUData(ud int, tname string) unsafe.Pointer {
	cs := C.CString(tname)
	defer C.free(unsafe.Pointer(cs))
	return C.luaL_checkudata(l.luaState, C.int(ud), cs)
}

func (l *LuaState) CheckVersion() {
	C.luaL_checkversion_(l.luaState, LUA_VERSION_NUM, LUAL_NUMSIZES)
}

func (l *LuaState) DoFile(path string) bool {
	return l.LoadFile(path) && l.PCall(0, LUA_MULTRET, 0)
}

func (l *LuaState) DoString(src string) bool {
	return l.LoadString(src) && l.PCall(0, LUA_MULTRET, 0)
}

func (l *LuaState) LError() {
	panic("not implemented, variadic function")
}

func (l *LuaState) ExecResult(stat int) int {
	return int(C.luaL_execresult(l.luaState, C.int(stat)))
}

func (l *LuaState) FileResult(stat int, fname string) int {
	cs := C.CString(fname)
	defer C.free(unsafe.Pointer(cs))
	return int(C.luaL_fileresult(l.luaState, C.int(stat), cs))
}

func (l *LuaState) GetMetaField(obj int, e string) bool {
	cs := C.CString(e)
	defer C.free(unsafe.Pointer(cs))
	return C.luaL_getmetafield(l.luaState, C.int(obj), cs) == LUA_OK
}

func (l *LuaState) LGetMetaTable(name string) bool {
	return l.GetField(LUA_REGISTRYINDEX, name)
}

func (l *LuaState) GetSubTable(idx int, fname string) bool {
	cs := C.CString(fname)
	defer C.free(unsafe.Pointer(cs))
	// If 0 then table was just created
	return C.luaL_getsubtable(l.luaState, C.int(idx), cs) != 0
}

func (l *LuaState) GSub(s, p, r string) string {
	cs := C.CString(s)
	cp := C.CString(p)
	cr := C.CString(r)
	defer C.free(unsafe.Pointer(cs))
	defer C.free(unsafe.Pointer(cp))
	defer C.free(unsafe.Pointer(cr))
	return C.GoString(C.luaL_gsub(l.luaState, cs, cp, cr))
}

func (l *LuaState) LLen(idx int) int {
	return int(C.luaL_len(l.luaState, C.int(idx)))
}

func (l *LuaState) LoadBuffer(buff []byte, name string) bool {
	return l.LoadBufferX(buff, name, "")
}

func (l *LuaState) LoadBufferX(buff []byte, name, mode string) bool {
	var cn *C.char = nil
	var cm *C.char = nil
	defer C.free(unsafe.Pointer(cn))
	defer C.free(unsafe.Pointer(cm))
	if len(name) > 0 {
		cn = C.CString(name)
	}
	if len(mode) > 0 {
		cm = C.CString(mode)
	}
	return C.luaL_loadbufferx(l.luaState, (*C.char)(unsafe.Pointer(&buff[0])), C.size_t(len(buff)), cn, cm) == LUA_OK
}

func (l *LuaState) LoadFile(path string) bool {
	return l.LoadFileX(path, "")
}

func (l *LuaState) LoadFileX(path, mode string) bool {
	ps := C.CString(path)
	defer C.free(unsafe.Pointer(ps))
	if len(mode) > 0 {
		cm := C.CString(mode)
		defer C.free(unsafe.Pointer(cm))
		return C.luaL_loadfilex(l.luaState, ps, cm) == LUA_OK
	} else {
		return C.luaL_loadfilex(l.luaState, ps, nil) == LUA_OK
	}
}

func (l *LuaState) LoadString(str string) bool {
	cs := C.CString(str)
	defer C.free(unsafe.Pointer(cs))
	return C.luaL_loadstring(l.luaState, cs) == LUA_OK
}

func (l *LuaState) NewLib() {
	//l.CheckVersion()
	//l.NewLibTable(l)
	//l.SetFuncs(l, 0)
	panic("not implemented")
}

func (l *LuaState) NewLibTable() {
	//C.lua_createtable(l.luaState, 0, sizeof(l)/sizeof((l)[0]) - 1)
	panic("not implemented")
}

func (l *LuaState) NewMetaTable(tname string) bool {
	cs := C.CString(tname)
	defer C.free(unsafe.Pointer(cs))
	return C.luaL_newmetatable(l.luaState, cs) == LUA_OK
}

func (l *LuaState) OpenLibs() {
	C.luaL_openlibs(l.luaState)
}

/*
luaL_opt
luaL_optinteger
luaL_optlstring
luaL_optnumber
luaL_optstring
luaL_prepbuffer
luaL_prepbuffsize
luaL_pushfail
luaL_pushresult
luaL_pushresultsize
luaL_ref
luaL_requiref
luaL_setfuncs
luaL_setmetatable
luaL_testudata
luaL_tolstring
luaL_traceback
luaL_typeerror
luaL_typename
luaL_unref
luaL_where

luaopen_base
luaopen_coroutine
luaopen_debug
luaopen_io
luaopen_math
luaopen_os
luaopen_package
luaopen_string
luaopen_table
luaopen_utf8
*/

///////////////////////////////////////////////////////////////////////////////
// Custom helper functions
///////////////////////////////////////////////////////////////////////////////

func (l *LuaState) ToInt(idx int) int {
	value := l.ToNumber(idx)
	if value < math.MinInt {
		return math.MinInt
	} else if value > math.MaxInt {
		return math.MaxInt
	} else {
		return int(value)
	}
}

func (l *LuaState) ToFloat32(idx int) float32 {
	value := l.ToNumber(idx)
	if value < math.SmallestNonzeroFloat32 {
		return math.SmallestNonzeroFloat32
	} else if value > math.MaxFloat32 {
		return math.MaxFloat32
	} else {
		return float32(value)
	}
}

func (l *LuaState) PushEnumKey(key string, val int) {
	l.PushString(key)
	l.PushInteger(int64(val))
	l.SetTable(-3)
}

//export print_stack
func print_stack(l *C.lua_State) C.int {
	L := luaMap[l]
	msg := L.ToLString(1)
	if len(msg) == 0 { // is error object not a string?
		// does it have a metamethod that produces a string?
		if L.CallMeta(1, "__tostring") && L.Type(-1) == LUA_TSTRING {
			return 1 // that is the message
		} else {
			str := fmt.Sprintf("(error object is a %s value)", L.TypeName(L.Type(1)))
			msg = L.PushString(str)
		}
	}
	cmsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cmsg))
	C.luaL_traceback(l, l, cmsg, 1) // append a standard traceback
	return 1                        // return the traceback
}

func (l *LuaState) CallSafe(nargs, nresults int) bool {
	base := l.GetTop() - nargs
	l.PushCFunction((C.lua_CFunction)(C.print_stack))
	l.Insert(base)
	ok := l.PCall(nargs, nresults, base)
	if !ok {
		log.Fatalf("%s", l.ToString(-1))
	}
	l.Remove(base)
	return ok
}

func (l *LuaState) Buffer(buff []byte) bool {
	if !l.LoadBuffer(buff, "") {
		log.Fatalf("error loading bytecode -> %s\n", l.ToString(-1))
		return false
	}
	return l.CallSafe(0, LUA_MULTRET)
}

func read_file(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buff, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (l *LuaState) BufferFile(file string) bool {
	buff, err := read_file(file)
	if err != nil {
		log.Fatalf("error reading file %s -> %s\n", file, err)
		return false
	}
	return l.Buffer(buff)
}

func (l *LuaState) BufferModule(moduleName string, buff []byte) bool {
	top := l.GetTop()
	l.GetGlobal("package")
	l.GetField(-1, "preload")
	success := false
	if !l.LoadBuffer(buff, "") {
		log.Fatalf("error loading bytecode -> %s\n", l.ToString(-1))
		success = false
	} else {
		l.SetField(-2, moduleName)
		success = true
	}
	l.SetTop(top)
	return success
}

func (l *LuaState) BufferModuleFile(moduleName, file string) bool {
	buff, err := read_file(file)
	if err != nil {
		log.Fatalf("error reading file %s -> %s\n", file, err)
		return false
	}
	return l.BufferModule(moduleName, buff)
}

func (l *LuaState) ModuleFile(moduleName, file string) bool {
	top := l.GetTop()
	l.GetGlobal("package")
	l.GetField(-1, "preload")
	if !l.LoadFile(file) {
		log.Fatalf("error loading bytecode -> %s", l.ToString(-1))
		return false
	} else {
		l.SetField(-2, moduleName)
	}
	l.SetTop(top)
	return true
}

func (l *LuaState) TryLoadFile(luaScript string) int {
	if l.LoadFile(luaScript) {
		return l.GetTop()
	}
	return -1
}

func (l *LuaState) TryLoadString(luaCode string) int {
	if l.LoadString(luaCode) {
		return l.GetTop()
	}
	return -1
}

func (l *LuaState) DoFileSafe(luaFile string) bool {
	if l.TryLoadFile(luaFile) != -1 {
		return l.CallSafe(0, LUA_MULTRET)
	} else {
		return false
	}
}

func (l *LuaState) DoStringSafe(luaCode string) bool {
	if !l.DoString(luaCode) {
		log.Fatalf("Failed to execute the string: %s", l.ToString(-1))
		l.Pop(1)
		return false
	} else {
		return true
	}
}

func (l *LuaState) ModuleFunc(module, function string) {
	l.GetGlobal("require")
	l.PushString(module)
	l.CallSafe(1, 1)
	if !l.IsTable(-1) {
		log.Fatalf("The module named %s could not be found\n", module)
	} else {
		l.TableFunc(function)
	}
}

func (l *LuaState) TableFunc(function string) {
	l.GetField(-1, function)
	if !l.IsFunction(-1) {
		log.Fatalf("The function named %s was not found in the table\n", function)
	} else {
		//lua_pushcfunction(lua, (lua_CFunction)local_print_stack);
		//lua_pushvalue(lua, -2);	// Copy luaClass on top of the stack
		//lua_pcall(lua, 0, 0, -2);
		//if (lua_pcall(lua, 0, 0, -2) != 0)
		//	log_err("Error calling %s::%s\n", module, func);
	}
}

func (l *LuaState) SetGlobalFunction(name string, fn func(*LuaState) int) {
	l.PushFunction(fn)
	l.SetGlobal(name)
}

func (l *LuaState) ArrayLength(offset int) int {
	if l.IsTable(offset) {
		return int(l.RawLen(offset))
	} else {
		log.Fatal("Expected an array but a table was not found at the offset")
		return 0
	}
}

func (l *LuaState) FieldArrayLength(field string, offset int) int {
	l.GetField(offset, field)
	count := l.ArrayLength(-1)
	l.Pop(1)
	return count
}

func (l *LuaState) FieldUserData(field string, offset int) unsafe.Pointer {
	var result unsafe.Pointer = nil
	l.GetField(offset, field)
	if l.IsLightUserData(-1) {
		result = l.ToUserData(-1)
	} else {
		log.Fatal("There was an error reading the user data value")
	}
	l.Pop(1)
	return result
}

func (l *LuaState) FieldUserDataAddress(field string, offset int) unsafe.Pointer {
	var result unsafe.Pointer = nil
	l.GetField(offset, field)
	if l.IsLightUserData(-1) {
		result = l.ToUserDataAddress(-1)
	} else {
		log.Fatal("There was an error reading the user data value")
	}
	l.Pop(1)
	return result
}

func (l *LuaState) FieldBool(field string, alt bool, offset int) bool {
	result := alt
	l.GetField(offset, field)
	if l.IsBoolean(-1) {
		result = l.ToBoolean(-1)
	} else {
		log.Fatal("There was an error reading the boolean value")
	}
	l.Pop(1)
	return result
}

func (l *LuaState) FieldInt(field string, alt int, offset int) int {
	result := alt
	l.GetField(offset, field)
	if l.IsNumber(-1) {
		result = l.ToInt(-1)
	} else {
		log.Fatal("There was an error reading the integer value")
	}
	l.Pop(1)
	return result
}

type Float interface {
	float32 | float64
}

func (l *LuaState) FieldFloat32(field string, alt float32, offset int) float32 {
	result := alt
	l.GetField(offset, field)
	if l.IsNumber(-1) {
		result = l.ToFloat32(-1)
	} else {
		log.Fatal("There was an error reading the float value")
	}
	l.Pop(1)
	return result
}

func (l *LuaState) FieldFloat64(field string, alt float64, offset int) float64 {
	result := alt
	l.GetField(offset, field)
	if l.IsNumber(-1) {
		result = l.ToNumber(-1)
	} else {
		log.Fatal("There was an error reading the float value")
	}
	l.Pop(1)
	return result
}

func (l *LuaState) FieldString(field string, alt string, offset int) string {
	result := alt
	l.GetField(offset, field)
	if l.IsString(-1) {
		result = l.ToString(-1)
	} else {
		log.Fatal("There was an error reading the string value")
	}
	l.Pop(1)
	return result
}

func (l *LuaState) CallFunc(nargs, nresults int) {
	l.CallSafe(nargs, nresults)
}
