package serializer_test

import (
	"testing"

	"github.com/flyingmt/pcbook/pb"
	"github.com/flyingmt/pcbook/sample"
	"github.com/flyingmt/pcbook/serializer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSerializer(t *testing.T) {
    t.Parallel()

    binaryFile := "../tmp/laptop.bin"
    jsonFile := "../tmp/laptop.json"

    laptop1 := sample.NewLaptop()
    err := serializer.WriteProtobuffToBinaryfile(laptop1, binaryFile)
    require.NoError(t, err)

    laptop2 := &pb.Laptop{}
    err = serializer.ReadProtobuffFromBinaryFile(binaryFile, laptop2)
    require.NoError(t, err)
    require.True(t, proto.Equal(laptop1, laptop2))

    err = serializer.WriteProtobuffToJSONFile(laptop1, jsonFile)
    require.NoError(t, err)

}
