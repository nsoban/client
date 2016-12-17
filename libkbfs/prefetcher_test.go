package libkbfs

import (
	"context"
	"testing"

	"github.com/keybase/go-codec/codec"
	"github.com/keybase/kbfs/kbfscodec"
	"github.com/keybase/kbfs/kbfshash"
	"github.com/stretchr/testify/require"
)

func makeFakeIndirectFilePtr(t *testing.T, off int64) IndirectFilePtr {
	return IndirectFilePtr{
		makeFakeBlockInfo(t),
		off,
		false,
		codec.UnknownFieldSetHandler{},
	}
}

func TestPrefetcherIndirectFileBlock(t *testing.T) {
	t.Log("Test indirect file block prefetching.")
	cache := NewBlockCacheStandard(10, getDefaultCleanBlockCacheCapacity())
	q := newBlockRetrievalQueue(1, kbfscodec.NewMsgpack(), cache)
	require.NotNil(t, q)
	defer q.Shutdown()

	bg := newFakeBlockGetter()
	w := newBlockRetrievalWorker(bg, q)
	require.NotNil(t, w)
	defer w.Shutdown()

	p := newPrefetcher(q)
	require.NotNil(t, q)
	q.prefetcher = p

	t.Log("Initialize an indirect file block pointing to 2 file data blocks.")
	ptrs := []IndirectFilePtr{
		makeFakeIndirectFilePtr(t, 0),
		makeFakeIndirectFilePtr(t, 150),
	}
	ptr1 := makeFakeBlockPointer(t)
	block1 := makeFakeFileBlock(t)
	block1.IPtrs = ptrs
	block1.IsInd = true
	block2 := makeFakeFileBlock(t)
	_, hash2 := kbfshash.DoRawDefaultHash(block2.Contents)
	block2.hash = &hash2
	block3 := makeFakeFileBlock(t)
	_, hash3 := kbfshash.DoRawDefaultHash(block3.Contents)
	block3.hash = &hash3

	_, continueCh1 := bg.setBlockToReturn(ptr1, block1)
	_, continueCh2 := bg.setBlockToReturn(ptrs[0].BlockPointer, block2)
	_, continueCh3 := bg.setBlockToReturn(ptrs[1].BlockPointer, block3)

	var block Block = &FileBlock{}
	ch := q.Request(context.Background(), defaultOnDemandRequestPriority, makeKMD(), ptr1, block, TransientEntry)
	continueCh1 <- struct{}{}
	err := <-ch
	require.NoError(t, err)
	require.Equal(t, block1, block)

	t.Log("Shutdown the prefetcher and wait until it's done prefetching.")
	go func() {
		continueCh2 <- struct{}{}
	}()
	go func() {
		continueCh3 <- struct{}{}
	}()
	<-p.Shutdown()

	t.Log("Ensure the prefetched blocks are in the cache.")
	block, err = cache.Get(ptr1)
	require.NoError(t, err)
	require.Equal(t, block1, block)
	block, err = cache.Get(ptrs[0].BlockPointer)
	require.NoError(t, err)
	require.Equal(t, block2, block)
	block, err = cache.Get(ptrs[1].BlockPointer)
	require.NoError(t, err)
	require.Equal(t, block3, block)
}
