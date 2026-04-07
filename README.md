# WhatsApp API Integration Sample

curl sample

```bash
curl -i -X POST \
  https://graph.facebook.com/{VERSION}/{PHONE_NUMBER_ID}/messages \
  -H 'Authorization: Bearer {WHATSAPP_ACCESS_TOKEN}' \
  -H 'Content-Type: application/json' \
  -d '{ "messaging_product": "whatsapp", "to": "{TO_PHONE_NUMBER}", "type": "template", "template": { "name": "hello_world", "language": { "code": "en_US" } } }'
```
