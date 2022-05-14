local get_dataset = std.native('get_dataset');
local get_instance_id = std.native('get_instance_id');
local get_one_of = std.native('get_one_of');
local get_timestamp = std.native('get_timestamp');
local get_integer = std.native('get_integer');
local get_number = std.native('get_number');
local get_rand_data = std.native('get_rand_data');
local get_rand_user_agent = std.native('get_rand_user_agent');

{
    id: get_integer(1, 1000),
    ip: get_rand_data()["ipv4"],
    user_agent: get_rand_user_agent(),
    prase: get_one_of("hi, moo, dooh, oops, whenever"),
}