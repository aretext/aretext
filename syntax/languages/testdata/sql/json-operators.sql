-- 13. Query with JSON operations
SELECT 
  user_id,
  profile_data->>'name' as name,
  profile_data->>'email' as email,
  profile_data->'address'->>'city' as city,
  profile_data->'address'->>'country' as country,
  jsonb_array_length(profile_data->'interests') as interest_count,
  jsonb_extract_path_text(profile_data, 'preferences', 'theme') as theme,
  profile_data->'subscription'->>'status' = 'active' as is_active_subscriber
FROM users
WHERE profile_data @> '{"subscription": {"plan": "premium"}}'
  AND profile_data ? 'verified'
  AND profile_data->'metrics'->>'visits' IS NOT NULL
ORDER BY (profile_data->'metrics'->>'visits')::integer DESC;
