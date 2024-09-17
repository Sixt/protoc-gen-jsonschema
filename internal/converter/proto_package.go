package converter

import (
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ProtoPackage describes a package of Protobuf, which is an container of message types.
type ProtoPackage struct {
	name     string
	parent   *ProtoPackage
	children map[string]*ProtoPackage
	types    map[string]*descriptorpb.DescriptorProto
}

func (c *Converter) lookupType(pkg *ProtoPackage, name string) (*descriptorpb.DescriptorProto, string, bool) {
	if strings.HasPrefix(name, ".") {
		return c.relativelyLookupType(globalPkg, name[1:len(name)])
	}

	for pkg != nil {
		if desc, pkgName, ok := c.relativelyLookupType(pkg, name); ok {
			return desc, pkgName, ok
		}

		pkg = pkg.parent
	}
	return nil, "", false
}

func (c *Converter) relativelyLookupType(pkg *ProtoPackage, name string) (*descriptorpb.DescriptorProto, string, bool) {
	head, tail, _ := strings.Cut(name, ".")

	if tail == "" {
		found, ok := pkg.types[head]
		return found, pkg.name, ok
	}

	c.Logger.Tracef("Looking for %s in %s at %s (%v)", tail, head, pkg.name, pkg)

	if child := pkg.children[head]; child != nil {
		found, pkgName, ok := c.relativelyLookupType(child, tail)
		return found, pkgName, ok
	}

	if msg := pkg.types[head]; msg != nil {
		found, ok := c.relativelyLookupNestedType(msg, tail)
		return found, pkg.name, ok
	}

	c.Logger.WithFields(logrus.Fields{
		"component": head,
		"package_name": pkg.name,
	}).Info("No such package nor message in package")

	return nil, "", false
}

func (c *Converter) relativelyLookupPackage(pkg *ProtoPackage, name string) (*ProtoPackage, bool) {
	for _, c := range strings.Split(name, ".") {
		next := pkg.children[c]
		if next == nil {
			return nil, false
		}

		pkg = next
	}

	return pkg, true
}
