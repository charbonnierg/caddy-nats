package azutils_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAzutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Azutils Suite")
}
