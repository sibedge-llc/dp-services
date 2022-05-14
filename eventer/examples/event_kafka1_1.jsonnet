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
    time: get_timestamp("2021-12-01", "2021-12-31", "1h"),
    group_id: get_instance_id("dummy_instance_id"),
    dataset: get_dataset("dummy_dataset"),
    country: get_one_of("ru, us, jp, cn"),
    name: get_rand_data()["first_name"],
    session_id: get_integer(1, 300),
    page_id: get_integer(1, 15),
    user_agent: get_rand_user_agent(),
    event: {
        type: get_one_of("click, open_page, button_click"),
        data: {
            [if self.type == "click" then self.type else null]: {
                x: get_number(1, 1000),
                y: get_number(1, 1000),
            },
            [if self.type == "button_click" then self.type else null]: {
                button_id: get_integer(1, 100),
            },
            [if self.type == "open_page" then self.type else null]: {
                page_id: $["page_id"],
            },
           
        },
    },
}