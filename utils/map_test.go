package utils

import (
	"reflect"
	"testing"
)

func TestMapByKey(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	type args struct {
		base    []User
		keyFunc func(User) int
	}
	tests := []struct {
		name string
		args args
		want map[int]User
	}{
		{
			name: "按ID建立映射",
			args: args{
				base: []User{
					{ID: 1, Name: "Alice"},
					{ID: 2, Name: "Bob"},
					{ID: 3, Name: "Charlie"},
				},
				keyFunc: func(u User) int {
					return u.ID
				},
			},
			want: map[int]User{
				1: {ID: 1, Name: "Alice"},
				2: {ID: 2, Name: "Bob"},
				3: {ID: 3, Name: "Charlie"},
			},
		},
		{
			name: "空切片",
			args: args{
				base: []User{},
				keyFunc: func(u User) int {
					return u.ID
				},
			},
			want: map[int]User{},
		},
		{
			name: "重复key后者覆盖前者",
			args: args{
				base: []User{
					{ID: 1, Name: "Alice"},
					{ID: 2, Name: "Bob"},
					{ID: 1, Name: "Alice2"},
				},
				keyFunc: func(u User) int {
					return u.ID
				},
			},
			want: map[int]User{
				1: {ID: 1, Name: "Alice2"},
				2: {ID: 2, Name: "Bob"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapByKey(tt.args.base, tt.args.keyFunc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapByKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapColumn(t *testing.T) {
	type User struct {
		ID   int
		Name string
		Age  int
	}

	type args struct {
		slice     []User
		extractor func(User) any
	}
	tests := []struct {
		name string
		args args
		want []any
	}{
		{
			name: "提取ID列",
			args: args{
				slice: []User{
					{ID: 1, Name: "Alice", Age: 20},
					{ID: 2, Name: "Bob", Age: 25},
					{ID: 3, Name: "Charlie", Age: 30},
				},
				extractor: func(u User) any {
					return u.ID
				},
			},
			want: []any{1, 2, 3},
		},
		{
			name: "提取Name列",
			args: args{
				slice: []User{
					{ID: 1, Name: "Alice", Age: 20},
					{ID: 2, Name: "Bob", Age: 25},
					{ID: 3, Name: "Charlie", Age: 30},
				},
				extractor: func(u User) any {
					return u.Name
				},
			},
			want: []any{"Alice", "Bob", "Charlie"},
		},
		{
			name: "空切片",
			args: args{
				slice: []User{},
				extractor: func(u User) any {
					return u.ID
				},
			},
			want: []any{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapColumn(tt.args.slice, tt.args.extractor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}
