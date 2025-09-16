package fakecapi

func (f *FakeCAPI) AddBroker(broker ServiceBroker, opts ...func(*FakeCAPI, ServiceBroker)) {
	if broker.Name == "" {
		broker.Name = fakeName("broker")
	}
	if broker.GUID == "" {
		broker.GUID = stableGUID(broker.Name)
	}

	f.brokers[broker.GUID] = broker
	for _, opt := range opts {
		opt(f, broker)
	}
}

type ServiceBroker struct {
	Name string
	GUID string
}
