package incr

import (
	"fmt"
	"strings"
)

// Dump returns a string for a given DAG.
func Dump(i Stabilizer) string {
	return dump(i, 0)
}

func dump(i Stabilizer, depth int) string {
	prefix := strings.Repeat("\t", depth)

	in := i.getNode()
	istr := fmt.Sprintf("%T", i)
	if len(in.parents) == 0 {
		return prefix + istr
	}

	sb := new(strings.Builder)
	sb.WriteString(prefix + istr)
	sb.WriteString(" {\n")
	for _, p := range in.parents {
		sb.WriteString(dump(p.getNode().self, depth+1))
		sb.WriteString(",\n")
	}
	sb.WriteString(prefix + "}")
	return sb.String()
}
