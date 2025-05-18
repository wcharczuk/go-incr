package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_ExpertScope_graph(t *testing.T) {
	g := New()
	testutil.Equal(t, true, ExpertScope(g).IsTopScope())
	testutil.Equal(t, true, ExpertScope(g).IsScopeNecessary())
	testutil.Equal(t, true, ExpertScope(g).IsScopeValid())

	testutil.Equal(t, -1, ExpertScope(g).ScopeHeight())
	testutil.NotNil(t, ExpertScope(g).ScopeGraph())
	testutil.Equal(t, g.ID(), ExpertScope(g).ScopeGraph().ID())
	testutil.NotNil(t, ExpertScope(g).NewIdentifier())

	/* nop */
	ExpertScope(g).AddScopeNode(new(mockBareNode))
}
