package incr

import "context"

// BindUpdate is a helper for dealing with bind node changes
// specifically handling unlinking and linking bound nodes
// when the bind changes.
func BindUpdate[A any](ctx context.Context, b Binder[A]) error {
	oldValue, newValue, err := b.Bind(ctx)
	if err != nil {
		return err
	}
	if oldValue == nil {
		Link(newValue, b)
		discoverAllNodes(ctx, b.Node().gs, newValue)
		b.SetBind(newValue)
		return nil
	}

	if oldValue.Node().id != newValue.Node().id {
		Unlink(oldValue)
		undiscoverAllNodes(ctx, b.Node().gs, oldValue)
		Link(newValue, b)
		discoverAllNodes(ctx, b.Node().gs, newValue)
		b.SetBind(newValue)
	}
	return nil
}
