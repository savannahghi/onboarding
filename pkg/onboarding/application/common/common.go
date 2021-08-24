package common

// AuthorizedEmails represent emails to check whether they have access to certain dto
var AuthorizedEmails = []string{"apa-dev@healthcloud.co.ke", "apa-prod@healthcloud.co.ke"}

// AuthorizedPhones represent phonenumbers to check whether they have access to certain dto
var AuthorizedPhones = []string{"+254700000000"}

// PubSub topic names
const (
	// CreateCustomerTopic is the TopicID for customer creation Topic
	CreateCustomerTopic = "customers.create"

	// CreateSupplierTopic is the TopicID for supplier creation Topic
	CreateSupplierTopic = "suppliers.create"

	// CreateCRMContact is the TopicID for CRM contact creation
	CreateCRMContact = "crm.contact.create"

	// UpdateCRMContact is the topicID for CRM contact updates
	UpdateCRMContact = "crm.contact.update"

	// LinkCoverTopic is the topicID for cover linking topic
	LinkCoverTopic = "covers.link"

	// LinkEDIMemberCoverTopic is the topic ID for cover linking topic of an EDI member who has
	// received a message with the link to download bewell
	LinkEDIMemberCoverTopic = "edi.covers.link"
)
