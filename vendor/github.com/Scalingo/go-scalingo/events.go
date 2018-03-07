package scalingo

import "gopkg.in/errgo.v1"

type EventsService interface {
	EventsList(app string, opts PaginationOpts) (Events, PaginationMeta, error)
	UserEventsList(opts PaginationOpts) (Events, PaginationMeta, error)
}

type EventsClient struct {
	subresourceClient
}

type EventsRes struct {
	Events []*Event `json:"events"`
	Meta   struct {
		PaginationMeta PaginationMeta `json:"pagination"`
	}
}

func (c *EventsClient) EventsList(app string, opts PaginationOpts) (Events, PaginationMeta, error) {
	var eventsRes EventsRes
	err := c.subresourceList(app, "events", opts.ToMap(), &eventsRes)
	if err != nil {
		return nil, PaginationMeta{}, errgo.Mask(err)
	}
	var events Events
	for _, ev := range eventsRes.Events {
		events = append(events, ev.Specialize())
	}
	return events, eventsRes.Meta.PaginationMeta, nil
}

func (c *EventsClient) UserEventsList(opts PaginationOpts) (Events, PaginationMeta, error) {
	req := &APIRequest{
		Client:   c.subresourceClient.backendConfiguration,
		Endpoint: "/events",
		Params:   opts.ToMap(),
	}

	var eventsRes EventsRes
	res, err := req.Do()
	if err != nil {
		return nil, PaginationMeta{}, errgo.Mask(err, errgo.Any)
	}

	err = ParseJSON(res, &eventsRes)
	if err != nil {
		return nil, PaginationMeta{}, errgo.Mask(err, errgo.Any)
	}

	var events Events
	for _, ev := range eventsRes.Events {
		events = append(events, ev.Specialize())
	}
	return events, eventsRes.Meta.PaginationMeta, nil
}
