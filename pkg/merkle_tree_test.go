package pkg

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var (
	defaultHashAlgo = Hash("sha256")

	ctx                = context.Background()
	configWithHashPool = MerkleTreeConfig{Hasher: &Hasher{
		Hash: defaultHashAlgo,
		Pool: NewHashPool(defaultHashAlgo.Hash()),
	}, MaxGoroutine: 1000}
	configWithNoHashPool = MerkleTreeConfig{Hasher: &Hasher{
		Hash: defaultHashAlgo,
		Pool: nil,
	}, MaxGoroutine: 1000}

	n1000000 = 1000000
	n100000  = 100000
	n1000    = 1000

	// generate leaf nodes data set
	dataNil         []Data
	dataEmpty       = []Data{}
	dataEvenNbNodes = []Data{
		StringData{Value: "value1"},
		StringData{Value: "value2"},
		StringData{Value: "value3"},
		StringData{Value: "value4"},
		StringData{Value: "value5"},
		StringData{Value: "value6"},
	}
	expectedEvenNbNodes = []*Node{
		{
			Data: dataEvenNbNodes[0],
		},
		{
			Data: dataEvenNbNodes[1],
		},
		{
			Data: dataEvenNbNodes[2],
		},
		{
			Data: dataEvenNbNodes[3],
		},
		{
			Data: dataEvenNbNodes[4],
		},
		{
			Data: dataEvenNbNodes[5],
		},
	}
	dataUnEvenNbNodes = []Data{
		StringData{Value: "value1"},
		StringData{Value: "value2"},
		StringData{Value: "value3"},
		StringData{Value: "value4"},
		StringData{Value: "value5"},
	}
	expectedUnEvenNbNodes = []*Node{
		{
			Data: dataUnEvenNbNodes[0],
		},
		{
			Data: dataUnEvenNbNodes[1],
		},
		{
			Data: dataUnEvenNbNodes[2],
		},
		{
			Data: dataUnEvenNbNodes[3],
		},
		{
			Data: dataUnEvenNbNodes[4],
		},
		{
			Data: dataUnEvenNbNodes[4],
		},
	}

	// generate parent nodes data set
	nodesNil []*Node

	mtWithEvenData, _   = NewMerkleTreeBuilder().WithHasher(configWithHashPool.Hasher).WithMaxGoroutine(configWithHashPool.MaxGoroutine).Build(ctx, dataEvenNbNodes)
	mtWithUnEvenData, _ = NewMerkleTreeBuilder().WithHasher(configWithHashPool.Hasher).WithMaxGoroutine(configWithHashPool.MaxGoroutine).Build(ctx, dataUnEvenNbNodes)
)

func TestMerkleTree_generateLeafNodes(t *testing.T) {
	type fields struct {
		Root             *Node
		Leaves           []*Node
		MerkleTreeConfig MerkleTreeConfig
	}
	type args struct {
		ctx  context.Context
		data []Data
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*Node
		err    error
	}{
		{
			"build merkle tree leaf nodes with nil data should return error",
			fields{
				Root:             nil,
				Leaves:           nil,
				MerkleTreeConfig: configWithHashPool,
			},
			args{
				ctx:  ctx,
				data: dataNil,
			},
			nil,
			ErrMerkleTreeDataIsNilOrEmpty,
		},
		{
			"build merkle tree leaf nodes with no data should return error",
			fields{
				Root:             nil,
				Leaves:           nil,
				MerkleTreeConfig: configWithHashPool,
			},
			args{
				ctx:  ctx,
				data: dataEmpty,
			},
			nil,
			ErrMerkleTreeDataIsNilOrEmpty,
		},
		{
			"build merkle tree leaf nodes with even nb of nodes should build",
			fields{
				Root:             nil,
				Leaves:           nil,
				MerkleTreeConfig: configWithHashPool,
			},
			args{
				ctx:  ctx,
				data: dataEvenNbNodes,
			},
			expectedEvenNbNodes,
			nil,
		},
		{
			"build merkle tree leaf nodes with uneven nb of nodes should build",
			fields{
				Root:             nil,
				Leaves:           nil,
				MerkleTreeConfig: configWithHashPool,
			},
			args{
				ctx:  ctx,
				data: dataUnEvenNbNodes,
			},
			expectedUnEvenNbNodes,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := &MerkleTree{
				Root:             tt.fields.Root,
				Leaves:           tt.fields.Leaves,
				MerkleTreeConfig: tt.fields.MerkleTreeConfig,
			}
			got, err := mt.generateLeafNodes(tt.args.ctx, tt.args.data)
			if tt.err != err {
				t.Errorf("generateLeafNodes() error = %v, wantErr %v", err, tt.err)
				return
			}
			for i, node := range got {
				if node.Data.String() != tt.want[i].Data.String() {
					t.Errorf("node.Data.String() is = %s, expected %s", node.Data.String(), tt.want[i].Data.String())
					return
				}
			}
		})
	}
}

func TestMerkleTree_generateParentNodes(t *testing.T) {
	type fields struct {
		Root             *Node
		Leaves           []*Node
		MerkleTreeConfig MerkleTreeConfig
	}
	type args struct {
		ctx       context.Context
		leafNodes []*Node
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Node
		err    error
	}{
		{
			name: "build merkle tree parent nodes with nil nodes should return error",
			fields: fields{
				Root:             nil,
				Leaves:           nil,
				MerkleTreeConfig: configWithHashPool,
			},
			args: args{
				ctx:       ctx,
				leafNodes: nodesNil,
			},
			want: nil,
			err:  ErrMerkleTreeDataIsNilOrEmpty,
		},
		{
			name: "build merkle tree parent nodes with empty nodes should return error",
			fields: fields{
				Root:             nil,
				Leaves:           nil,
				MerkleTreeConfig: configWithHashPool,
			},
			args: args{
				ctx:       ctx,
				leafNodes: nodesNil,
			},
			want: nil,
			err:  ErrMerkleTreeDataIsNilOrEmpty,
		},
		{
			name: "build merkle tree parent nodes with even nodes should return merkle tree root",
			fields: fields{
				Root:             nil,
				Leaves:           nil,
				MerkleTreeConfig: configWithHashPool,
			},
			args: args{
				ctx:       ctx,
				leafNodes: mtWithEvenData.Leaves,
			},
			want: mtWithEvenData.Root,
			err:  nil,
		},
		{
			name: "build merkle tree parent nodes with uneven nodes should return merkle tree root",
			fields: fields{
				Root:             nil,
				Leaves:           nil,
				MerkleTreeConfig: configWithHashPool,
			},
			args: args{
				ctx:       ctx,
				leafNodes: mtWithUnEvenData.Leaves,
			},
			want: mtWithUnEvenData.Root,
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := &MerkleTree{
				Root:             tt.fields.Root,
				Leaves:           tt.fields.Leaves,
				MerkleTreeConfig: tt.fields.MerkleTreeConfig,
			}
			got, err := mt.generateParentNodes(tt.args.ctx, tt.args.leafNodes)
			if tt.err != err {
				t.Errorf("generateParentNodes() error = %v, wantErr %v", err, tt.err)
				return
			}
			if tt.want != nil && !reflect.DeepEqual(got.Hash, tt.want.Hash) {
				t.Errorf("generateParentNodes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerkleTree_Verify(t *testing.T) {
	type fields struct {
		Root             *Node
		Leaves           []*Node
		MerkleTreeConfig MerkleTreeConfig
	}
	type args struct {
		data Data
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
		err    error
	}{
		{
			name: "build an empty tree should return false",
			fields: fields{
				Root:             nil,
				Leaves:           nil,
				MerkleTreeConfig: configWithHashPool,
			},
			args: args{
				data: dataEvenNbNodes[0],
			},
			want: false,
			err:  nil,
		},
		{
			name: "build a data with reused buffer leaf should return true",
			fields: fields{
				Root:             mtWithEvenData.Root,
				Leaves:           mtWithEvenData.Leaves,
				MerkleTreeConfig: configWithHashPool,
			},
			args: args{
				data: dataEvenNbNodes[0],
			},
			want: true,
			err:  nil,
		},
		{
			name: "build a data without reused buffer leaf should return true",
			fields: fields{
				Root:             mtWithEvenData.Root,
				Leaves:           mtWithEvenData.Leaves,
				MerkleTreeConfig: configWithNoHashPool,
			},
			args: args{
				data: dataEvenNbNodes[0],
			},
			want: true,
			err:  nil,
		},
		{
			name: "build a data leaf that is not present in the tree should return false",
			fields: fields{
				Root:             mtWithEvenData.Root,
				Leaves:           mtWithEvenData.Leaves,
				MerkleTreeConfig: configWithHashPool,
			},
			args: args{
				data: StringData{
					Value: "not=present",
				},
			},
			want: false,
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := &MerkleTree{
				Root:             tt.fields.Root,
				Leaves:           tt.fields.Leaves,
				MerkleTreeConfig: tt.fields.MerkleTreeConfig,
			}
			got, err := mt.Verify(ctx, tt.args.data)
			if tt.err != err {
				t.Errorf("generateParentNodes() error = %v, wantErr %v", err, tt.err)
				return
			}
			if got != tt.want {
				t.Errorf("Verify() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkMerkleTreeBuilder_Build_N1000(b *testing.B) {
	build(b, n1000)
}
func BenchmarkMerkleTreeBuilder_Build_N100000(b *testing.B) {
	build(b, n100000)
}
func BenchmarkMerkleTreeBuilder_Build_N1000000(b *testing.B) {
	build(b, n1000000)
}

func BenchmarkMerkleTreeBuilder_Verify_N1000(b *testing.B) {
	verify(b, n1000)
}

func BenchmarkMerkleTreeBuilder_Verify_N100000(b *testing.B) {
	verify(b, n100000)
}

func BenchmarkMerkleTreeBuilder_Verify_N1000000(b *testing.B) {
	verify(b, n1000000)
}

func build(b *testing.B, n int) {
	data := make([]Data, n)
	for i := 0; i < n; i++ {
		data[i] = StringData{Value: fmt.Sprintf("valueeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee%d", i)}
	}
	for i := 0; i < b.N; i++ {
		_, err := NewMerkleTreeBuilder().WithHasher(configWithHashPool.Hasher).WithMaxGoroutine(configWithHashPool.MaxGoroutine).Build(ctx, data)
		if err != nil {
			log.Error(err)
		}
		assert.NoError(b, err)
	}
}

func verify(b *testing.B, n int) {
	data := make([]Data, n)
	for i := 0; i < n; i++ {
		data[i] = StringData{Value: fmt.Sprintf("valueeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee%d", i)}
	}
	mt, err := NewMerkleTreeBuilder().WithHasher(configWithHashPool.Hasher).WithMaxGoroutine(configWithHashPool.MaxGoroutine).Build(ctx, data)
	assert.NoError(b, err)
	for i := 0; i < b.N; i++ {
		var isTrue bool
		isTrue, err = mt.Verify(ctx, data[n/2-1])
		if err != nil {
			log.Error(err)
			assert.Error(b, err)
		}
		assert.NoError(b, err)
		assert.Equal(b, true, isTrue)
	}
}
