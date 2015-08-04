package types

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/attic-labs/noms/chunks"
	"github.com/attic-labs/noms/ref"
	"github.com/stretchr/testify/assert"
)

func getTestCompoundBlob(datas ...string) compoundBlob {
	blobs := make([]Future, len(datas))
	childLengths := make([]uint64, len(datas))
	length := uint64(0)
	for i, s := range datas {
		b, _ := NewBlob(bytes.NewBufferString(s))
		blobs[i] = futureFromValue(b)
		childLengths[i] = uint64(len(s))
		length += uint64(len(s))
	}
	return compoundBlob{length, childLengths, blobs, &ref.Ref{}, nil}
}

func getAliceBlob(t *testing.T) compoundBlob {
	assert := assert.New(t)
	f, err := os.Open("alice-short.txt")
	assert.NoError(err)
	defer f.Close()

	b, err := NewBlob(f)
	assert.NoError(err)
	cb, ok := b.(compoundBlob)
	assert.True(ok)
	return cb
}

func TestCompoundBlobReader(t *testing.T) {
	assert := assert.New(t)
	cb := getTestCompoundBlob("hello", "world")
	bs, err := ioutil.ReadAll(cb.Reader())
	assert.NoError(err)
	assert.Equal("helloworld", string(bs))

	ab := getAliceBlob(t)
	bs, err = ioutil.ReadAll(ab.Reader())
	assert.NoError(err)
	f, err := os.Open("alice-short.txt")
	assert.NoError(err)
	defer f.Close()
	bs2, err := ioutil.ReadAll(f)
	assert.Equal(bs2, bs)
}

func TestCompoundBlobLen(t *testing.T) {
	assert := assert.New(t)
	cb := getTestCompoundBlob("hello", "world")
	assert.Equal(uint64(10), cb.Len())

	ab := getAliceBlob(t)
	assert.Equal(uint64(30157), ab.Len())
}

func TestCompoundBlobChunks(t *testing.T) {
	assert := assert.New(t)
	cs := &chunks.MemoryStore{}

	cb := getTestCompoundBlob("hello", "world")
	assert.Equal(0, len(cb.Chunks()))

	bl1 := newBlobLeaf([]byte("hello"))
	blr1 := bl1.Ref()
	bl2 := newBlobLeaf([]byte("world"))
	cb = compoundBlob{uint64(10), []uint64{5, 5}, []Future{futureFromRef(blr1), futureFromValue(bl2)}, &ref.Ref{}, cs}
	assert.Equal(1, len(cb.Chunks()))
}

func TestCompoundBlobSameChunksWithPrefix(t *testing.T) {
	assert := assert.New(t)

	cb1 := getAliceBlob(t)

	// Load same file again but prepend some data... all but the first chunk should stay the same
	f, err := os.Open("alice-short.txt")
	assert.NoError(err)
	defer f.Close()
	buf := bytes.NewBufferString("prefix")
	r := io.MultiReader(buf, f)

	b, err := NewBlob(r)
	assert.NoError(err)
	cb2 := b.(compoundBlob)

	assert.Equal(cb2.Len(), cb1.Len()+uint64(6))
	assert.Equal(3, len(cb1.blobs))
	assert.Equal(len(cb1.blobs), len(cb2.blobs))
	assert.NotEqual(cb1.blobs[0].Ref(), cb2.blobs[0].Ref())
	assert.Equal(cb1.blobs[1].Ref(), cb2.blobs[1].Ref())
	assert.Equal(cb1.blobs[2].Ref(), cb2.blobs[2].Ref())
}

func TestCompoundBlobSameChunksWithSuffix(t *testing.T) {
	assert := assert.New(t)

	cb1 := getAliceBlob(t)

	// Load same file again but append some data... all but the last chunk should stay the same
	f, err := os.Open("alice-short.txt")
	assert.NoError(err)
	defer f.Close()
	buf := bytes.NewBufferString("suffix")
	r := io.MultiReader(f, buf)

	b, err := NewBlob(r)
	assert.NoError(err)
	cb2 := b.(compoundBlob)

	assert.Equal(cb2.Len(), cb1.Len()+uint64(6))
	assert.Equal(3, len(cb1.blobs))
	assert.Equal(len(cb1.blobs), len(cb2.blobs))
	assert.Equal(cb1.blobs[0].Ref(), cb2.blobs[0].Ref())
	assert.Equal(cb1.blobs[1].Ref(), cb2.blobs[1].Ref())
	assert.NotEqual(cb1.blobs[2].Ref(), cb2.blobs[2].Ref())
}
