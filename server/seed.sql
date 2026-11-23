-- Devices
INSERT INTO devices (name, type, driver, external_id) VALUES
  ('Desk lamp', 'light', 'tuya', 'bf9855e603861192e8agws'),
  ('Test lamp', 'light', 'mock', 'test-1')
ON CONFLICT DO NOTHING;

-- Scenes (device_id 1 = Desk lamp)
-- "temp" = real white LEDs (0 warm - 100 cool); "color" = RGB LEDs. Use one or the other.
INSERT INTO scenes (name, actions) VALUES
  ('study', '[{"device_id": 1, "power": true, "temp": 70, "brightness": 100}]'),
  ('chill', '[{"device_id": 1, "power": true, "temp": 10, "brightness": 40}]'),
  ('sleep', '[{"device_id": 1, "power": false}]')
ON CONFLICT (name) DO UPDATE SET actions = EXCLUDED.actions;