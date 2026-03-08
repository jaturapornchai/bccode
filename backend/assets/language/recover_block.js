const fs = require('fs');

const thData = JSON.parse(fs.readFileSync('temp_th.json', 'utf8'));

const startKey = "no_products_found";
const endKey = "please_specify_option_group_name";

const recovered = {};

// Sort thData by code just in case, though it seems sorted or random?
// The array in temp_th.json seems somewhat random or grouped?
// "code": "code", "menu_master", "address", "all"... alphabetical?
// "code", "menu_master", "address" -> "a" comes after "m"? No.
// So it is NOT sorted alphabetically.
// But languages.json IS sorted alphabetically.

// So I just need to find ALL codes that fall in the alphabetical range.

for (const item of thData) {
    if (item.code >= startKey && item.code <= endKey) {
        recovered[item.code] = {
            "cn": item.text, // Placeholder
            "en": item.text, // Placeholder
            "ja": item.text, // Placeholder
            "km": item.text, // Placeholder
            "ko": item.text, // Placeholder
            "lo": item.text, // Placeholder
            "my": item.text, // Placeholder
            "th": item.text,
            "vi": item.text  // Placeholder
        };
    }
}

// Sort the recovered keys alphabetically to match languages.json structure
const sortedRecovered = {};
Object.keys(recovered).sort().forEach(key => {
    sortedRecovered[key] = recovered[key];
});

// Output as a JSON string without the outer braces, so I can paste it in?
// Or just full JSON.
fs.writeFileSync('recovered_block.json', JSON.stringify(sortedRecovered, null, 2));
