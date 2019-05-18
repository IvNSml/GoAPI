package main

import (
	"fmt"
	"math"
)

type PosPeaks struct {
	Pos   []int
	Peaks []int
}
func main() {
	var  a1 [] int= nil
	a2 := []int{11*21, 121*121, 144*144, 19*19, 161*161, 19*19, 144*144, 19*19}
	fmt.Println(Comp(a1,a2))
}
func Comp(array1 []int, array2 []int) bool {
	var sqrt int
	notfirst:=false
	if array2==nil || array1==nil || len(array2)==0 || len(array1)==0{
		return false
	}
	for _,val1:=range array1{
		sqrt=int(math.Sqrt(float64(val1)))
		for _,val2:=range array2{
			if sqrt==val2{
				break
			}else if val2==array2[len(array2)-1]{
				 notfirst=true
				 break
			}
		}
		if notfirst{
			break
		}
		if val1==array1[len(array1)-1]{
			return true
		}

	}
	for _,val2:=range array2{
		sqrt=int(math.Sqrt(float64(val2)))
		for _,val1:=range array1{
			if sqrt==val1{
				break
			}else if val1==array1[len(array1)-1]{
				return false
			}
		}
		if val2==array2[len(array2)-1]{
		return true
		}
	}
	return false
}

