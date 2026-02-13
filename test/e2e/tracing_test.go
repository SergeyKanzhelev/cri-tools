/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "go.opentelemetry.io/proto/otlp/trace/v1"
)

var _ = t.Describe("tracing", func() {
	It("should generate spans for version command", func() {
		t.CrictlWithTracing("version")
		// It might fail with permission denied if not run as root, 
		// but it should still generate a span if tracing is enabled.
		
		// Wait a bit for spans to be exported (crictl flushes on exit)
		Eventually(func() int {
			return len(t.OtelCollector.GetSpans())
		}, "5s", "100ms").Should(BeNumerically(">", 0))

		spans := t.OtelCollector.GetSpans()
		
		foundRoot := false
		for _, span := range spans {
			if span.Name == "version" {
				foundRoot = true
				break
			}
		}
		Expect(foundRoot).To(BeTrue(), "Root span 'version' not found")
	})

	It("should generate child spans for pods command", func() {
		// This requires sudo to actually talk to the socket and generate child spans
		res := t.CrictlWithTracing("pods")
		
		Eventually(func() int {
			return len(t.OtelCollector.GetSpans())
		}, "5s", "100ms").Should(BeNumerically(">", 0))

		spans := t.OtelCollector.GetSpans()
		
		var rootSpan *v1.Span
		var childSpan *v1.Span
		for _, span := range spans {
			if span.Name == "pods" {
				rootSpan = span
			}
			// Look for CRI gRPC child spans
			if strings.Contains(span.Name, "ListPodSandbox") {
				childSpan = span
			}
		}
		Expect(rootSpan).NotTo(BeNil(), "Root span 'pods' not found")
		
		if res.ExitCode() == 0 {
			Expect(childSpan).NotTo(BeNil(), "Child span 'ListPodSandbox' not found")
			Expect(bytes.Equal(rootSpan.TraceId, childSpan.TraceId)).To(BeTrue(), 
				"Root span and child span should have the same trace ID")
			Expect(bytes.Equal(childSpan.ParentSpanId, rootSpan.SpanId)).To(BeTrue(),
				"Child span should have root span as parent")
		}
	})

	It("should generate Version gRPC span as a child of the root span", func() {
		res := t.CrictlWithTracing("version")
		
		Eventually(func() int {
			return len(t.OtelCollector.GetSpans())
		}, "5s", "100ms").Should(BeNumerically(">", 0))

		spans := t.OtelCollector.GetSpans()
		
		var rootSpan *v1.Span
		var versionChildSpan *v1.Span
		for _, span := range spans {
			if span.Name == "version" {
				rootSpan = span
			}
			// The gRPC span name for Version call
			if strings.Contains(span.Name, "runtime.v1.RuntimeService/Version") {
				versionChildSpan = span
			}
		}
		
		Expect(rootSpan).NotTo(BeNil(), "Root span 'version' not found")
		
		// If sudo was used and connection succeeded, we should have the child span
		if res.ExitCode() == 0 {
			Expect(versionChildSpan).NotTo(BeNil(), "Child span for Version API not found")
			Expect(bytes.Equal(rootSpan.TraceId, versionChildSpan.TraceId)).To(BeTrue(), 
				"Root span and Version child span should have the same trace ID")
			Expect(bytes.Equal(versionChildSpan.ParentSpanId, rootSpan.SpanId)).To(BeTrue(),
				"Version child span should have root span as parent")
		}
	})
})
