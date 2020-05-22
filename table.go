package lua

import "fmt"

const defaultArrayCap = 4
const defaultHashCap = 4

/*
type lValueArraySorter struct {
	L      *LState
	Fn     *LFunction
	Values []LValue
}

func (lv lValueArraySorter) Len() int {
	return len(lv.Values)
}

func (lv lValueArraySorter) Swap(i, j int) {
	lv.Values[i], lv.Values[j] = lv.Values[j], lv.Values[i]
}

func (lv lValueArraySorter) Less(i, j int) bool {
	if lv.Fn != nil {
		lv.L.Push(lv.Fn)
		lv.L.Push(lv.Values[i])
		lv.L.Push(lv.Values[j])
		lv.L.Call(2, 1)
		return LVAsBool(lv.L.reg.Pop())
	}
	return lessThan(lv.L, lv.Values[i], lv.Values[j])
}
*/

func newLTable(acap int, hcap int) *LTable {
	if acap < 0 {
		acap = 0
	}
	if hcap < 0 {
		hcap = 0
	}
	tb := &LTable{}
	tb.Metatable = LNil
	initltable(&tb.tab, acap+hcap)
	return tb
}

// Len returns length of this LTable.
func (tb *LTable) Len() int {
	/*
		if tb.array == nil {
			return 0
		}
		var prev LValue = LNil
		for i := len(tb.array) - 1; i >= 0; i-- {
			v := tb.array[i]
			if prev == LNil && v != LNil {
				return i + 1
			}
			prev = v
		}
	*/
	return int(tb.tab.GetN())
}

// Append appends a given LValue to this LTable.
func (tb *LTable) Append(value LValue) {
	/*
		if value == LNil {
			return
		}
		if tb.array == nil {
			tb.array = make([]LValue, 0, defaultArrayCap)
		}
		if len(tb.array) == 0 || tb.array[len(tb.array)-1] != LNil {
			tb.array = append(tb.array, value)
		} else {
			i := len(tb.array) - 2
			for ; i >= 0; i-- {
				if tb.array[i] != LNil {
					break
				}
			}
			tb.array[i+1] = value
		}
	*/
	tb.Insert(-1, value)
}

// Insert inserts a given LValue at position `i` in this table.
func (tb *LTable) Insert(pos int, value LValue) {
	e := int(tb.tab.GetN()) + 1
	if pos == -1 {
		pos = e
	}
	if !(1 <= pos && pos <= e) {
		panic("position out of bounds")
	}
	i := 0
	for i = e; i > pos; i-- { /* move up elements */
		pv := *tb.tab.GetInt(int64(i) - 1)
		_ = tb.tab.SetInt(int64(i), pv)
	}
	_ = tb.tab.SetInt(int64(pos), value)
}

// MaxN returns a maximum number key that nil value does not exist before it.
func (tb *LTable) MaxN() int {
	k, _, ok := tb.tab.Next(LNil)
	max := LNumber(0)
	for ok {
		if kv, ok := k.(LNumber); ok {
			if kv > max {
				max = kv
			}
		}
		k, _, ok = tb.tab.Next(k)
	}
	return int(max)

	/*
		if tb.array == nil {
			return 0
		}
		for i := len(tb.array) - 1; i >= 0; i-- {
			if tb.array[i] != LNil {
				return i + 1
			}
		}
		return 0
	*/
}

// Remove removes from this table the element at a given position.
func (tb *LTable) Remove(pos int) LValue {
	size := int(tb.tab.GetN())
	if pos == -1 {
		pos = size
	}
	if pos != size {
		if !(1 <= pos && pos <= size+1) {
			panic("position out of bounds")
		}
	}
	oldval := *tb.tab.GetInt(int64(pos))
	for ; pos < size; pos++ {
		nv := *tb.tab.GetInt(int64(pos) + 1)
		_ = tb.tab.SetInt(int64(pos), nv)
	}
	_ = tb.tab.SetInt(int64(pos), LNil)
	fmt.Println(tb.tab.array, tb.tab.node)
	return oldval
}

// RawSet sets a given LValue to a given index without the __newindex metamethod.
// It is recommended to use `RawSetString` or `RawSetInt` for performance
// if you already know the given LValue is a string or number.
func (tb *LTable) RawSet(key LValue, value LValue) {
	p, err := tb.tab.Set(key)
	if err != nil {
		panic(err)
	}
	*p = value
}

// RawSetInt sets a given LValue at a position `key` without the __newindex metamethod.
func (tb *LTable) RawSetInt(key int, value LValue) {
	err := tb.tab.SetInt(int64(key), value)
	if err != nil {
		panic(err)
	}
}

// RawSetString sets a given LValue to a given string index without the __newindex metamethod.
func (tb *LTable) RawSetString(key string, value LValue) {
	tb.RawSet(LString(key), value)
}

// RawSetH sets a given LValue to a given index without the __newindex metamethod.
func (tb *LTable) RawSetH(key LValue, value LValue) {
	tb.RawSet(key, value)
}

// RawGet returns an LValue associated with a given key without __index metamethod.
func (tb *LTable) RawGet(key LValue) LValue {
	return *tb.tab.Get(key)
}

// RawGetInt returns an LValue at position `key` without __index metamethod.
func (tb *LTable) RawGetInt(key int) LValue {
	return *tb.tab.GetInt(int64(key))
}

// RawGet returns an LValue associated with a given key without __index metamethod.
func (tb *LTable) RawGetH(key LValue) LValue {
	return *tb.tab.Get(key)
}

// RawGetString returns an LValue associated with a given key without __index metamethod.
func (tb *LTable) RawGetString(key string) LValue {
	return *tb.tab.Get(LString(key))
}

// ForEach iterates over this table of elements, yielding each in turn to a given function.
func (tb *LTable) ForEach(cb func(LValue, LValue)) {
	k, v, ok := tb.tab.Next(LNil)
	for ok {
		cb(k, v)
		k, v, ok = tb.tab.Next(k)
	}
}

// This function is equivalent to lua_next ( http://www.lua.org/manual/5.1/manual.html#lua_next ).
func (tb *LTable) Next(key LValue) (LValue, LValue) {
	k, v, _ := tb.tab.Next(key)
	return k, v
}
