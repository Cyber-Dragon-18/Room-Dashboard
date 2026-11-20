-- Devices

INSERT INTO devices (name, type, driver, external_id) VALUES
  ('Desk lamp', 'light', 'tuya', 'bf9855e603861192e8agws'),
  ('Test lamp', 'light', 'mock', 'test-1')
ON CONFLICT DO NOTHING;


-- Scenes
INSERT INTO scenes (name, actions) VALUES
  ('study', '[{"device_id": 1, "power": true, "color": {"r": 255, "g": 255, "b": 255}, "brightness": 100}]'),
  ('chill', '[{"device_id": 1, "power": true, "color": {"r": 255, "g": 120, "b": 40},  "brightness": 40}]'),
  ('sleep', '[{"device_id": 1, "power": false}]')
ON CONFLICT DO NOTHING;    