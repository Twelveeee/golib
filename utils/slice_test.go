package utils

import (
	"errors"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

func TestForEach(t *testing.T) {
	type args struct {
		data []int
		f    func(int) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "遍历不报错",
			args: args{
				data: []int{1, 2, 3, 4},
				f:    func(int) error { return nil },
			},
		}, {
			name: "遍历不报错",
			args: args{
				data: []int{1, 2, 3, 4},
				f: func(i int) error {
					if i > 2 {
						return errors.New("大于2")
					}
					return nil
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ForEach(tt.args.data, tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("ForEach() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindIndex(t *testing.T) {
	type args struct {
		data []int
		f    func(int) bool
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "找不到",
			args: args{
				data: []int{1, 2, 3, 4},
				f: func(i int) bool {
					return i > 10
				},
			},
			want: -1,
		}, {
			name: "找到了",
			args: args{
				data: []int{1, 2, 3, 4},
				f: func(i int) bool {
					return i == 3
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindIndex(tt.args.data, tt.args.f); got != tt.want {
				t.Errorf("FindIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindItem(t *testing.T) {
	type args struct {
		data   []int
		target int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "找不到",
			args: args{
				data:   []int{1, 2, 3, 4, 5},
				target: 8,
			},
			want: -1,
		}, {
			name: "找到了",
			args: args{
				data:   []int{1, 2, 3, 4, 5},
				target: 1,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindItem(tt.args.data, tt.args.target); got != tt.want {
				t.Errorf("FindItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMap(t *testing.T) {
	type args struct {
		data []int
		f    func(int) string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "+1后String",
			args: args{
				data: []int{1, 2, 3},
				f: func(i int) string {
					return strconv.FormatInt(int64(i+1), 10)
				},
			},
			want: []string{
				"2", "3", "4",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Map(tt.args.data, tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	type args struct {
		data []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			args: args{
				data: []int{
					3, 2, 6, 2, 3, 1,
				},
			},
			want: []int{ // unique无序,结果排序比较
				1, 2, 3, 6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Unique(tt.args.data)
			sort.Ints(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsIn(t *testing.T) {
	type args struct {
		data   []int
		target int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "in",
			args: args{
				data:   []int{1, 2, 3, 4},
				target: 2,
			},
			want: true,
		}, {
			name: "not in",
			args: args{
				data:   []int{1, 2, 3, 4},
				target: 8,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InArray(tt.args.target, tt.args.data); got != tt.want {
				t.Errorf("IsIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	type args struct {
		data []int
		f    func(int) bool
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			args: args{
				data: []int{1, 2, 3, 4, 5},
				f: func(i int) bool {
					return i%2 == 0
				},
			},
			want: []int{2, 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filter(tt.args.data, tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestChunk(t *testing.T) {
	type args struct {
		data []int
		size int
	}
	tests := []struct {
		name string
		args args
		want [][]int
	}{
		{
			name: "good size",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				size: 5,
			},
			want: [][]int{
				{1, 2, 3, 4, 5},
				{6, 7, 8, 9, 10},
			},
		}, {
			name: "last not full",
			args: args{
				data: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				size: 3,
			},
			want: [][]int{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
				{10},
			},
		}, {
			name: "big size",
			args: args{
				data: []int{1, 2, 3},
				size: 5,
			},
			want: [][]int{
				{1, 2, 3},
			},
		},
	}
	for _, tt := range tests {
		if tt.name != "last not full" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := Chunk(tt.args.data, tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Slice() = %v, want %v", got, tt.want)
			}
		})
	}
}
