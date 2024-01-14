package nonce

import "golang.org/x/exp/constraints"

type List[T constraints.Ordered] []T

func (l List[T]) Len() int { return len(l) }

func (l List[T]) Less(i, j int) bool { return l[i] < l[j] }

func (l List[T]) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

func (l *List[T]) Push(x any) { *l = append(*l, x.(T)) }

func (l *List[T]) Pop() (x any) {
	x, *l = (*l)[len(*l)-1], (*l)[0:len(*l)-1]
	return
}
