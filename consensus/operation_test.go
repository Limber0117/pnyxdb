/**
 * Copyright (c) 2019 - Present – Thomson Licensing, SAS
 * All rights reserved.
 *
 * This source code is licensed under the Clear BSD license found in the
 * LICENSE file in the root directory of this source tree.
 */

package consensus

import (
	"fmt"
	"testing"

	"github.com/technicolor-research/pnyxdb/consensus/operations"

	"github.com/stretchr/testify/require"
)

func TestOperation_CheckConflict(t *testing.T) {
	ok := func(t *testing.T, a, b *Operation) {
		require.Nil(t, a.CheckConflict(b))
		require.Nil(t, b.CheckConflict(a))
	}
	ko := func(t *testing.T, a, b *Operation) {
		require.NotNil(t, a.CheckConflict(b))
		require.NotNil(t, b.CheckConflict(a))
	}

	t.Run("SET SET different key", func(t *testing.T) {
		op1 := &Operation{Key: "a", Op: Operation_SET, Data: []byte("hello")}
		op2 := &Operation{Key: "b", Op: Operation_SET, Data: []byte("world")}
		ok(t, op1, op2)
	})
	t.Run("SET SET same data", func(t *testing.T) {
		op1 := &Operation{Key: "c", Op: Operation_SET, Data: []byte("hello")}
		op2 := &Operation{Key: "c", Op: Operation_SET, Data: []byte("hello")}
		ok(t, op1, op2)
	})
	t.Run("SET SET", func(t *testing.T) {
		op1 := &Operation{Key: "d", Op: Operation_SET, Data: []byte("hello")}
		op2 := &Operation{Key: "d", Op: Operation_SET, Data: []byte("world")}
		ko(t, op1, op2)
	})
	t.Run("SET ADD", func(t *testing.T) {
		op1 := &Operation{Key: "e", Op: Operation_SET, Data: []byte{0x01}}
		op2 := &Operation{Key: "e", Op: Operation_ADD, Data: []byte{0x02}}
		ko(t, op1, op2)
	})
	t.Run("ADD ADD", func(t *testing.T) {
		op1 := &Operation{Key: "f", Op: Operation_ADD, Data: []byte{0x01}}
		op2 := &Operation{Key: "f", Op: Operation_ADD, Data: []byte{0x02}}
		ok(t, op1, op2)
	})
	t.Run("SADD SREM different data", func(t *testing.T) {
		op1 := &Operation{Key: "a", Op: Operation_SADD, Data: []byte("add")}
		op2 := &Operation{Key: "a", Op: Operation_SREM, Data: []byte("rem")}
		ok(t, op1, op2)
	})
	t.Run("SADD SREM same data", func(t *testing.T) {
		op1 := &Operation{Key: "a", Op: Operation_SADD, Data: []byte("hey")}
		op2 := &Operation{Key: "a", Op: Operation_SREM, Data: []byte("hey")}
		ko(t, op1, op2)
	})
}

func TestOperation_Exec_Simple(t *testing.T) {
	opSet := &Operation{Op: Operation_SET, Data: []byte("hello")}
	opAdd := &Operation{Op: Operation_ADD, Data: []byte("1.5")}
	opMul := &Operation{Op: Operation_MUL, Data: []byte("3")}
	opBad := &Operation{Op: Operation_MUL, Data: []byte("bad")}

	type execCase struct {
		op          *Operation
		data        []byte
		resExpected []byte
		errExpected bool
	}
	testCases := []execCase{
		{opSet, []byte("world"), []byte("hello"), false},
		{opSet, nil, []byte("hello"), false},
		{opAdd, []byte("2.5"), []byte("4"), false},
		{opMul, []byte("2.5"), []byte("7.5"), false},
		{opAdd, []byte{}, []byte("1.5"), false},
		{opMul, []byte{}, []byte("0"), false},
		{opAdd, nil, []byte("1.5"), false},
		{opMul, nil, []byte("0"), false},
		{opAdd, []byte("2.x"), nil, true},
		{opMul, []byte("2.x"), nil, true},
		{opBad, []byte("2.5"), nil, true},
	}

	for _, tc := range testCases {
		c := tc
		t.Run(fmt.Sprintf("%s/%s", c.op.Op.String(), c.data), func(t *testing.T) {
			value := operations.NewValue(c.data)
			err := c.op.Exec(value)
			if !c.errExpected {
				require.Nil(t, err)
				require.Exactly(t, c.resExpected, value.Raw)
			} else {
				require.NotNil(t, err)
			}
		})
	}
}
