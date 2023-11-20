const nats = await import("https://cdn.jsdelivr.net/npm/nats.ws@1.18.0/esm/nats.js")

// Create a nats client:
nc = await nats.connect(
	{"servers": "wss://peripheral1.local.quara-dev.com:10444", "user": "APP", "token": document.cookie}
)
