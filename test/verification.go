package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/ptypes/any"
	"net/http"
	"statefun-go/pkg/flink/statefun"
	"statefun-go/pkg/flink/statefun/io"
)

func randToken(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		panic("Fail :(")
	}
	return hex.EncodeToString(bytes)
}

type CounterFunction struct{}

func (c CounterFunction) Invoke(runtime statefun.StatefulFunctionRuntime, _ *any.Any) error {
	var count InvokeCount
	if err := runtime.Get("invoke_count", &count); err != nil {
		return fmt.Errorf("unable to deserialize invoke_count %w", err)
	}

	count.Count += 1

	if err := runtime.Set("invoke_count", &count); err != nil {
		return fmt.Errorf("unable to serialize invoke_count %w", err)
	}

	response := &InvokeResult{
		InvokeCount: count.Count,
		Id:          runtime.Self().Id,
	}

	target := statefun.Address{
		FunctionType: statefun.FunctionType{
			Namespace: "org.apache.flink.statefun.e2e.remote",
			Type:      "forward-function",
		},
		Id: randToken(0xFFFF),
	}

	return runtime.Send(target, response)
}

func ForwardFunction(runtime statefun.StatefulFunctionRuntime, msg *any.Any) error {
	egress := io.EgressIdentifier{
		EgressNamespace: "org.apache.flink.statefun.e2e.remote",
		EgressType:      "invoke-results",
	}

	record := io.KafkaRecord{
		Topic: "invoke-results",
		Key:   runtime.Self().Id,
		Value: msg,
	}

	message, err := record.ToMessage()
	if err != nil {
		return fmt.Errorf("unable to serialize message to kafka %w", err)
	}

	return runtime.SendEgress(egress, message)
}

func main() {
	registry := statefun.NewFunctionRegistry()

	registry.RegisterFunction(statefun.FunctionType{
		Namespace: "org.apache.flink.statefun.e2e.remote",
		Type:      "counter",
	}, CounterFunction{})

	registry.RegisterFunctionPointer(statefun.FunctionType{
		Namespace: "org.apache.flink.statefun.e2e.remote",
		Type:      "forward-function",
	}, ForwardFunction)

	http.Handle("/service", registry)
	_ = http.ListenAndServe(":8000", nil)
}
