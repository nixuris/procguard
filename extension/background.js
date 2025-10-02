// The name of the native messaging host.
const nativeHostName = "com.nixuris.procguard";

let port;

function connect() {
  console.log("Attempting to connect to native host:", nativeHostName);
  port = chrome.runtime.connectNative(nativeHostName);

  // Immediately ask for the blocklist upon connecting.
  port.postMessage({ command: "getBlocklist" });

  port.onMessage.addListener((message) => {
    console.log("Received message from native host:", message);
    if (message.domains) {
      updateBlockingRules(message.domains);
    }
  });

  port.onDisconnect.addListener(() => {
    if (chrome.runtime.lastError) {
      console.error("Disconnected with error:", chrome.runtime.lastError.message);
    }
    console.log("Disconnected from native host. Reconnecting in 5 seconds...");
    setTimeout(connect, 5000);
  });
}

async function updateBlockingRules(domains) {
  const rules = domains.map((domain, index) => ({
    id: index + 1,
    priority: 1,
    action: { type: "block" },
    condition: { urlFilter: `||${domain}` }
  }));

  try {
    const existingRules = await chrome.declarativeNetRequest.getDynamicRules();
    const existingRuleIds = existingRules.map(rule => rule.id);
    
    await chrome.declarativeNetRequest.updateDynamicRules({
      removeRuleIds: existingRuleIds,
      addRules: rules
    });
    console.log("Updated blocking rules for domains:", domains);
  } catch (error) {
    console.error("Error updating blocking rules:", error);
  }
}

// Initial connection attempt
connect();
