package container_test

import (
	"testing"

	"github.com/liuliqiang/log4go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/liuliqiang/cloud-native-demo/internal/service/container"
)

func TestDockerListInstance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Container Operator Test")
}

var _ = Describe("Docker container test", func() {
	dockerOper := container.NewDockerOper("0")

	Context("list container", func() {
		It("should return instances match labels", func() {
			insts, err := dockerOper.ListInstanceWithLabel("branch", "docker/golang-sdk")
			Expect(err).To(BeNil())
			for i := 0; i < len(insts); i++ {
				log4go.Info(insts[i].Name)
			}
			Expect(len(insts)).To(Equal(3))
		})
	})

})
