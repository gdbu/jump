package permissions

import core "github.com/gdbu/dbl"

type source interface {
	New(value core.Value) (resourceID string, err error)
	GetByRelationship(relationship, relationshipID string, value interface{}) error
}

func getByKey(s source, resourceKey string) (r *Resource, err error) {
	var rs []*Resource
	if err = s.GetByRelationship(relationshipResourceKeys, resourceKey, &rs); err != nil {
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
