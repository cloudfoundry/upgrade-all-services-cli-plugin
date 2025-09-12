package fakecapi

func WithServiceOffering(offering ServiceOffering, opts ...func(*FakeCAPI, ServiceOffering)) func(*FakeCAPI, ServiceBroker) {
	return func(f *FakeCAPI, broker ServiceBroker) {
		if offering.Name == "" {
			offering.Name = guid()
		}
		if offering.GUID == "" {
			offering.GUID = guid()
		}
		offering.ServiceBrokerName = broker.Name
		offering.ServiceBrokerGUID = broker.GUID

		f.offerings[offering.GUID] = offering

		for _, opt := range opts {
			opt(f, offering)
		}
	}
}

type ServiceOffering struct {
	Name              string `json:"name"`
	GUID              string `json:"guid"`
	ServiceBrokerName string `json:"-"`
	ServiceBrokerGUID string `json:"-"`
}
