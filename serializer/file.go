package serializer

import (
	"fmt"
	"io/ioutil"

	"google.golang.org/protobuf/proto"
)

func WriteProtobuffToJSONFile(message proto.Message, filename string) error {
    data, err := ProtobufToJSON(message)
    if err != nil {
        return fmt.Errorf("cannot marshal proto message to json: %w", err)
    }

    err = ioutil.WriteFile(filename, []byte(data), 0644)
    if err != nil {
        return fmt.Errorf("cannot write json data to file: %w", err)
    }

    return nil
}

// WriteProtobuffToBinaryFile writes protocol buffer message to binary file
func WriteProtobuffToBinaryfile(message proto.Message, filename string) error {
    data, err := proto.Marshal(message)
    if err != nil {
        return fmt.Errorf("cannot marshal proto message to binary: %w", err)
    }

    err = ioutil.WriteFile(filename, data, 0644)
    if err != nil {
        return fmt.Errorf("cannot write binrary data to file: %w", err)
    }

    return nil
}

// ReadProtobuffFromBinaryFile reads protocol buffer message from binary file
func ReadProtobuffFromBinaryFile(filename string, message proto.Message) error {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("cannot read binary data from file: %w", err)
    }

    err = proto.Unmarshal(data, message)
    if err != nil {
        return fmt.Errorf("cannot unmarshal binary to proto message: %w", err)
    }

    return nil
}
