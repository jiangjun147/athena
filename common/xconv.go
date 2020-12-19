package common

import (
	"reflect"
)

func getValueOfElem(val reflect.Value) reflect.Value {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

func xconv(srcVal reflect.Value, dstVal reflect.Value) interface{} {
	f := srcVal.MethodByName("Xconv")
	if f.IsValid() {
		rs := f.Call([]reflect.Value{})
		dstVal.Set(rs[0])
		return rs[0].Interface()
	}

	srcVal = srcVal.Elem()
	srcType := srcVal.Type()

	v := dstVal.Elem()
	for i := 0; i < srcType.NumField(); i++ {
		ft := srcType.Field(i)
		tag := ft.Tag.Get("xconv")

		fv := getValueOfElem(srcVal.Field(i))
		if !fv.IsValid() {
			continue
		}

		f := v.FieldByName(ft.Name)
		if !f.IsValid() {
			continue
		}

		switch tag {
		case "-":
		case "int":
			f.SetInt(fv.Int())
		case "glc_encode":
			f.SetString(GlcEncode(fv.Int()))
		case "":
			f.Set(fv)
		default:
			rs := fv.MethodByName(tag).Call([]reflect.Value{})
			f.Set(rs[0])
		}
	}

	return dstVal.Interface()
}

func xconvSlice(srcVal reflect.Value, dstVal reflect.Value) interface{} {
	srcLen := srcVal.Len()

	dstType := dstVal.Type()
	slice := reflect.MakeSlice(reflect.SliceOf(dstType), srcLen, srcLen)

	for i := 0; i < srcLen; i++ {
		dstElem := slice.Index(i)
		if dstElem.IsNil() {
			dstElem.Set(reflect.New(dstType.Elem()))
		}
		xconv(srcVal.Index(i), dstElem)
	}

	return slice.Interface()
}

// Xconv 不同结构体间拷贝
// src: *StructA
//   tag: "-" 不拷贝 "Func" 调用src.Func()转化
// dst: *StructB
// return: *StructB
func Xconv(src interface{}, dst interface{}) interface{} {
	if src == nil {
		return nil
	}

	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	if srcVal.Kind() == reflect.Slice {
		return xconvSlice(srcVal, dstVal)
	}

	return xconv(srcVal, dstVal)
}
