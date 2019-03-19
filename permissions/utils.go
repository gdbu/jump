package permissions

import core "github.com/Hatch1fy/service-core"

type source interface {
	New(value core.Value) (resourceID string, err error)
	GetByRelationship(relationship, relationshipID string, value interface{}) error
}

func getByKey(s source, resourceKey string) (ep *Entry, err error) {
	var es []*Entry
	if err = s.GetByRelationship(relationshipResourceKeys, resourceKey, &es); err != nil {
		return
	}

	if len(es) == 0 {
		err = ErrResourceNotFound
		return
	}

	ep = es[0]
	return
}

func getOrCreateByKey(s source, resourceKey string) (ep *Entry, err error) {
	if ep, err = getByKey(s, resourceKey); err != ErrResourceNotFound {
		return
	}

	e := newEntry(resourceKey)
	if _, err = s.New(&e); err != nil {
		return
	}

	ep = &e
	return
}
