package fakecapi

func (f *FakeCAPI) AddBroker(broker ServiceBroker, opts ...func(*FakeCAPI, ServiceBroker)) {
	if broker.Name == "" {
		broker.Name = guid()
	}
	if broker.GUID == "" {
		broker.GUID = guid()
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
