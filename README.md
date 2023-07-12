# Cgo-Lua
Lua 5.4 Cgo implementaiton for Go (golang) that always tries to stay up to date with the latest public release of the Lua library.

## Goal
Some people are not interested in being stuck on Lua 5.1 forever and wish to use the new language features. This is the primary reason for making this repository. This uses nearly identical Lua 5.4 code (with the exception of detecting linux and [setting a define](https://github.com/BrentFarris/Cgo-Lua/blob/master/luaconf.h#L8)).

## Build
I've tested this on Windows and Linux (not Mac yet, I don't own one). For the most part you just need to import the library and you're good to go. Though, in order to build the Lua source (part of this package) you'll need to make sure you have the following environment variable set:
```sh
CGO_LDFLAGS="-lm -ldl"
```

## Usage
Hello world:
```go
import (
	lua "github.com/BrentFarris/Cgo-Lua"
)

func main() {
	L := lua.NewLuaState()
	defer L.Close()
	L.GetGlobal("print")
	L.PushString("Hello, World! -From the Cgo API!")
	L.Call(1, 0)
}
```

Hello using go function:
```go
import (
	"log"
	lua "github.com/BrentFarris/Cgo-Lua"
)

func main() {
	L := lua.NewLuaState()
	L.SetGlobalFunction("other_hello", func(L *lua.LuaState) int {
		log.Print("Hello Go function from Lua!")
		return 0
	})
	L.DoString("other_hello()")
}
```

## Challenges
There are a couple of choices and challenges that were needed to be worked out to successfully bind the Lua library

### Calling Go functions from Lua
The primary problem is making it easy to pass a Go function for Lua to call. Typically you'd need to create a C function for every instance, this library is setup so that you can directly pass a Go function and have it call as you would expect without creating a C function. This uses a [dirty little hack](https://github.com/BrentFarris/Cgo-Lua/blob/master/wrapper.go#L418C6-L418C6) to make it possible, and by using a lookup table for your function ID. This hack basically has Lua push an anonomous function to the top of the stack to bind to.

### Passing Go pointers to Lua
Typically a pointer to a Go structure will have a pointer to another Go structure within it. Due to this, you can not pass a pointer to this object to Lua... or can you? If we do some dirty hacks where we turn the pointer into a number, you can pass this to Lua and have Lua pass it back.

**WARNING!!!** Be sure to keep a reference to that pointer somewhere in Go, you'll have a bad time if it is collected by the garbage collector while in Lua land. Tracking this is your responsibility.

You can use `lua.PushFunction` to pass a go function to Lua for calling, or you can use a standard C function with `lua.PushCFunction` if you'd rather not deal with the aformentioned hack.
