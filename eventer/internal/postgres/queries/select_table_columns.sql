SELECT 
   column_name,
   lower(data_type) as data_type,
   lower(udt_name) as udt_name,
   is_nullable
FROM 
   information_schema.columns
WHERE 
   table_name = :table_name
