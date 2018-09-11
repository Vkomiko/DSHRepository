package main

/*

*/
import "C"
import (
	"unsafe"
	"fmt"
	"reflect"
)

func func0() ([5]int, *[5]int, []int) {
	var arr0 = [5]int{1, 2, 3, 4, 5}
	var arr1 = [5]int{6, 7, 8, 9, 10}
	var arr2 = [5]int{11, 12, 13, 14, 15}

	fmt.Println("func0 &arr0", unsafe.Pointer(&arr0))
	fmt.Println("func0 &arr1", unsafe.Pointer(&arr1))
	fmt.Println("func0 &arr2 + 3*8", unsafe.Pointer(uintptr(unsafe.Pointer(&arr2)) + uintptr(3 * 8)))

	return arr0, // 传值
		//以下是安全的, 因为go有escape analysis, 编译时会自动分配在堆上
		&arr1,
		arr2[3: 5]
}

func main() {
	var arr0 [5]int = [5]int{1, 2, 3, 4, 5}

	//  fmt.Println(unsafe.Pointer(arr0))  //数组引用不是指针, 是常指针?
	fmt.Println(unsafe.Pointer(&arr0))
	fmt.Println(&(arr0[0]), &(arr0[1]), &(arr0[2]))

	var arr1 []int = make([]int, 5, 5)
	var arr2 *[]int = new([]int)
	var arr3 *[5]int = new([5]int)
	fmt.Println(arr1, arr2, arr3)

	r0, r1, r2 := func0()
	fmt.Println("func0 ret: ", r0, r1, r2)
	fmt.Println("func0 ret *: ",
		unsafe.Pointer(&r0),
		unsafe.Pointer(r1),
		unsafe.Pointer(&r2),
		// TODO: 切片可强行转换为指向头元素的指针, 这时go的特性还是hack?
		*(**int)(unsafe.Pointer(&r2)),
		unsafe.Pointer(&r2[0]))

	var p0 *int = (*int)(unsafe.Pointer(&arr0))
	// 指针转换为切片
	s0 := (*[5]int)(unsafe.Pointer(p0))[: 5: 5]
	fmt.Println("s0 type: ", reflect.TypeOf(s0))
	fmt.Println("s0: ", s0)

	// 切片转换为指针
	p1 := &s0[0]
	fmt.Println("p1 type: ", reflect.TypeOf(p1))
	fmt.Println("p0 == p1: ", p0 == p1)

	p2 := *(**int)(unsafe.Pointer(&s0))  // Hack ?
	fmt.Println("p0 == p2: ", p0 == p2)
}

