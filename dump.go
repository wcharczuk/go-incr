package incr

import (
	"fmt"
	"strings"
)

// Dump returns a string for a given DAG.
func Dump[A any](i Stabilizer) string {
	return dump[A](i, 0)
}

func dump[A any](i Stabilizer, depth int) string {
	prefix := strings.Repeat("\t", depth)
	valueProvider, _ := i.(Incr[A])

	in := i.getNode()
	istr := fmt.Sprintf("%T (%v)", i, valueProvider.Value())
	if len(in.parents) == 0 {
		return prefix + istr
	}

	sb := new(strings.Builder)
	sb.WriteString(prefix + istr)
	sb.WriteString(" {\n")
	for _, p := range in.parents {
		sb.WriteString(dump[A](p.getNode().self, depth+1))
		sb.WriteString(",\n")
	}
	sb.WriteString(prefix + "}")
	return sb.String()
}
