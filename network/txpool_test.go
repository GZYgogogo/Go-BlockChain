package network

import (
	"strconv"
	"testing"

	"projectx/core"

	"github.com/stretchr/testify/assert"
)

func TestNewTxPool(t *testing.T) {
	p := NewTxPool(10)
	assert.Equal(t, 0, p.PendingCount())
}

func TestTxPoolAddTx(t *testing.T) {
	p := NewTxPool(11)
	n := 10
	for i := 1; i <= n; i++ {
		tx := core.NewTransaction([]byte(strconv.FormatInt(int64(i), 10)))
		p.Add(tx)
		// the same txx that have same hash cannot add twice
		p.Add(tx)

		assert.Equal(t, i, p.PendingCount())
		assert.Equal(t, i, p.pending.Count())
		assert.Equal(t, i, p.all.Count())
	}
}

func TestSortTranscation(t *testing.T) {
	p := NewTxPool(11)
	n := 10

	for i := 1; i <= n; i++ {
		tx := core.NewTransaction([]byte(strconv.FormatInt(int64(i), 10)))
		p.Add(tx)
		// cannot add twice
		p.Add(tx)

		assert.Equal(t, i, p.PendingCount())
		assert.Equal(t, i, p.pending.Count())
		assert.Equal(t, i, p.all.Count())
	}
}

func TestTxPoolMaxLength(t *testing.T) {
	maxLen := 10
	p := NewTxPool(maxLen)
	n := 100
	txx := []*core.Transaction{}

	for i := 0; i < n; i++ {
		tx := core.NewTransaction([]byte(strconv.FormatInt(int64(i), 10)))
		p.Add(tx)

		if i > n-(maxLen+1) {
			txx = append(txx, tx)
		}
	}

	assert.Equal(t, p.all.Count(), maxLen)
	assert.Equal(t, len(txx), maxLen)

	for _, tx := range txx {
		assert.True(t, p.Contains(tx.Hash(core.TxHasher{})))
	}
}

func TestTxSortedMapFirst(t *testing.T) {
	m := NewTxSortedMap()
	first := core.NewTransaction([]byte(strconv.FormatInt(int64(1), 10)))
	m.Add(first)
	m.Add(core.NewTransaction([]byte(strconv.FormatInt(int64(2), 10))))
	m.Add(core.NewTransaction([]byte(strconv.FormatInt(int64(3), 10))))
	m.Add(core.NewTransaction([]byte(strconv.FormatInt(int64(4), 10))))
	m.Add(core.NewTransaction([]byte(strconv.FormatInt(int64(5), 10))))
	assert.Equal(t, first, m.First())
}

func TestTxSortedMapAdd(t *testing.T) {
	m := NewTxSortedMap()
	n := 100

	for i := 0; i < n; i++ {
		tx := core.NewTransaction([]byte(strconv.FormatInt(int64(i), 10)))
		m.Add(tx)
		// cannot add the same twice
		m.Add(tx)

		assert.Equal(t, m.Count(), i+1)
		assert.True(t, m.Contains(tx.Hash(core.TxHasher{})))
		assert.Equal(t, len(m.lookup), m.txx.Len())
		assert.Equal(t, m.Get(tx.Hash(core.TxHasher{})), tx)
	}

	m.Clear()
	assert.Equal(t, m.Count(), 0)
	assert.Equal(t, len(m.lookup), 0)
	assert.Equal(t, m.txx.Len(), 0)
}

func TestTxSortedMapRemove(t *testing.T) {
	m := NewTxSortedMap()

	tx := core.NewTransaction([]byte(strconv.FormatInt(int64(1), 10)))
	m.Add(tx)
	assert.Equal(t, m.Count(), 1)

	m.Remove(tx.Hash(core.TxHasher{}))
	assert.Equal(t, m.Count(), 0)
	assert.False(t, m.Contains(tx.Hash(core.TxHasher{})))
}
