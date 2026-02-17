const fs = require('fs');
const path = require('path');

console.log('\n========================================');
console.log('🔍 BUILD DIAGNOSTICS - CHECKING OUTPUT');
console.log('========================================\n');

const distPath = path.join(__dirname, 'dist_v3');

function listFilesRecursively(dir, indent = '') {
  try {
    const items = fs.readdirSync(dir);
    items.forEach(item => {
      const fullPath = path.join(dir, item);
      const stats = fs.statSync(fullPath);

      if (stats.isDirectory()) {
        console.log(`${indent}📁 ${item}/`);
        listFilesRecursively(fullPath, indent + '  ');
      } else {
        const size = (stats.size / 1024).toFixed(2);
        console.log(`${indent}📄 ${item} (${size} KB)`);
      }
    });
  } catch (err) {
    console.error(`❌ Error reading directory ${dir}:`, err.message);
  }
}

// Check if dist_v3 directory exists
if (fs.existsSync(distPath)) {
  console.log(`✅ Found dist_v3 directory at: ${distPath}\n`);
  console.log('📂 Directory structure:\n');
  listFilesRecursively(distPath);

  // Check for index.html specifically
  const indexPath = path.join(distPath, 'index.html');
  if (fs.existsSync(indexPath)) {
    console.log('\n✅ index.html found at root of dist_v3/');

    // Read first 500 chars of index.html to verify content
    const indexContent = fs.readFileSync(indexPath, 'utf-8');
    console.log('\n📝 First 500 characters of index.html:');
    console.log('---');
    console.log(indexContent.substring(0, 500));
    console.log('---\n');

    // Check for our VICTORY DEPLOY marker
    if (indexContent.includes('VICTORY DEPLOY')) {
      console.log('🎉 SUCCESS: Found "VICTORY DEPLOY" in index.html!');
    } else {
      console.log('⚠️  WARNING: "VICTORY DEPLOY" not found in index.html');
    }

    // Check for BUILD_ID marker
    if (indexContent.includes('VICTORY_DEPLOY_2126')) {
      console.log('🎉 SUCCESS: Found "VICTORY_DEPLOY_2126" build ID in index.html!');
    } else {
      console.log('⚠️  WARNING: "VICTORY_DEPLOY_2126" build ID not found in index.html');
    }
  } else {
    console.log('\n❌ index.html NOT found at root of dist_v3/');
  }
} else {
  console.log(`❌ dist_v3 directory NOT found at: ${distPath}`);
  console.log('⚠️  Build may have failed or output to different location');
}

console.log('\n========================================');
console.log('✅ DIAGNOSTICS COMPLETE');
console.log('========================================\n');
