package permissions

import (
	"github.com/mojura/mojura"
	"github.com/mojura/mojura/filters"
)

type source interface {
	New(value mojura.Value) (resourceID string, err error)
	GetFiltered(value interface{}, opts *mojura.FilteringOpts) (lastID string, err error)
}

func getByKey(s source, resourceKey string) (r *Resource, err error) {
	var rs []*Resource
	filter := filters.Match(relationshipResourceKeys, resourceKey)
	opts := mojura.NewFilteringOpts(filter)
	if _, err = s.GetFiltered(&rs, opts); err != nil {
		return
	}

	if len(rs) == 0 {
		err = ErrResourceNotFound
		return
	}

	r = rs[0]
	return
}

func getOrCreateByKey(s source, resourceKey string) (rp *Resource, err error) {
	if rp, err = getByKey(s, resourceKey); err != ErrResourceNotFound {
		return
	}

	r := newResource(resourceKey)
	if _, err = s.New(&r); err != nil {
		return
	}

	rp = &r
	return
}
