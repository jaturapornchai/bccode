const fs = require('fs');

const recovered = JSON.parse(fs.readFileSync('recovered_block.json', 'utf8'));
const languages = fs.readFileSync('languages.json', 'utf8').split('\n');

// 1-based lines from view_file: 12877 to 12889
// 0-based: 12876 to 12888
const startLine = 12876;
const endLine = 12888;

console.log(`Replacing lines ${startLine + 1} to ${endLine + 1}`);
console.log(`Start content: ${languages[startLine]}`);
console.log(`End content: ${languages[endLine]}`);

if (!languages[startLine].includes('"no_products_found": {')) {
    console.error('Start line mismatch!');
    process.exit(1);
}

// Prepare replacement content
let replacementLines = [];
const keys = Object.keys(recovered).sort();
for (let i = 0; i < keys.length; i++) {
    const key = keys[i];
    const val = recovered[key];
    replacementLines.push(`  "${key}": {`);
    const langKeys = Object.keys(val).sort();
    for (let j = 0; j < langKeys.length; j++) {
        const lk = langKeys[j];
        // Escape quotes in value if needed (though temp_th.json shouldn't have them unescaped)
        // But simple string concatenation might be risky if value has quotes.
        // JSON.stringify is safer for value.
        const valStr = JSON.stringify(val[lk]);
        const comma = j < langKeys.length - 1 ? ',' : '';
        replacementLines.push(`    "${lk}": ${valStr}${comma}`);
    }
    // Always add comma because 'please_wait' follows in the file
    replacementLines.push(`  },`);
}

// Replace
languages.splice(startLine, endLine - startLine + 1, ...replacementLines);

fs.writeFileSync('languages.json', languages.join('\n'));
console.log('Recovery applied successfully.');
