const net = require('net');
const { chromium, firefox, webkit } = require('playwright');

const DEFAULT_TIMEOUT = 30_000;

function parseConfig() {
  const raw = process.env.WEBAUTO_RUNNER_CONFIG;
  if (!raw) {
    throw new Error('WEBAUTO_RUNNER_CONFIG not provided');
  }

  try {
    const parsed = JSON.parse(raw);
    if (!parsed.browserType) {
      throw new Error('browserType is required');
    }

    return {
      browserType: parsed.browserType,
      headless: parsed.headless !== undefined ? Boolean(parsed.headless) : true,
    };
  } catch (error) {
    throw new Error(`Failed to parse runner config: ${error.message}`);
  }
}

function resolveBrowserLauncher(browserType) {
  switch (browserType) {
    case 'chromium':
      return chromium;
    case 'firefox':
      return firefox;
    case 'webkit':
      return webkit;
    default:
      throw new Error(`Invalid browser type: ${browserType}`);
  }
}

function toCommandError(error) {
  return JSON.stringify({
    success: false,
    error: error instanceof Error ? error.message : String(error),
  });
}

async function handleCommand(page, command) {
  const timeout = typeof command.timeout === 'number' ? command.timeout : DEFAULT_TIMEOUT;

  switch (command.command) {
    case 'navigate': {
      await page.goto(command.url, {
        waitUntil: command.waitUntil || 'load',
        timeout,
      });
      return {
        success: true,
        data: {
          url: page.url(),
          title: await page.title(),
        },
      };
    }

    case 'click': {
      const element = page.locator(command.selector);
      await element.click({ timeout });
      return {
        success: true,
        data: {
          selector: command.selector,
          clicked: true,
        },
      };
    }

    case 'screenshot': {
      const screenshot = await page.screenshot({
        type: command.type || 'png',
        fullPage: Boolean(command.fullPage),
        timeout,
      });
      return {
        success: true,
        data: {
          screenshot: screenshot.toString('base64'),
          type: command.type || 'png',
          fullPage: Boolean(command.fullPage),
        },
      };
    }

    case 'type': {
      const element = page.locator(command.selector);
      await element.fill(command.text, { timeout });
      return {
        success: true,
        data: {
          selector: command.selector,
          text: command.text,
          typed: true,
        },
      };
    }

    case 'pdf': {
      const pdf = await page.pdf({
        format: command.format || 'A4',
        landscape: Boolean(command.landscape),
        printBackground:
          command.printBackground !== undefined ? Boolean(command.printBackground) : true,
        timeout,
      });
      return {
        success: true,
        data: {
          pdf: pdf.toString('base64'),
          format: command.format || 'A4',
          landscape: Boolean(command.landscape),
          printBackground:
            command.printBackground !== undefined ? Boolean(command.printBackground) : true,
        },
      };
    }

    case 'get-text': {
      const element = page.locator(command.selector);
      const count = await element.count();
      if (count === 0) {
        throw new Error(`Element not found: ${command.selector}`);
      }

      let text;
      if (count === 1) {
        text = await element.textContent({ timeout });
      } else {
        text = await element.allTextContents();
      }

      return {
        success: true,
        data: {
          selector: command.selector,
          text,
          element_count: count,
        },
      };
    }

    case 'get-attribute': {
      const element = page.locator(command.selector);
      const count = await element.count();
      if (count === 0) {
        throw new Error(`Element not found: ${command.selector}`);
      }

      let attributeValue;
      if (count === 1) {
        attributeValue = await element.getAttribute(command.attributeName, { timeout });
      } else {
        const values = [];
        for (let i = 0; i < count; i += 1) {
          values.push(await element.nth(i).getAttribute(command.attributeName));
        }
        attributeValue = values;
      }

      return {
        success: true,
        data: {
          selector: command.selector,
          attribute_name: command.attributeName,
          attribute_value: attributeValue,
          element_count: count,
        },
      };
    }

    case 'wait': {
      const element = page.locator(command.selector);
      const startTime = Date.now();

      await element.waitFor({
        state: command.waitCondition || 'visible',
        timeout,
      });

      const waitedMs = Date.now() - startTime;
      const count = await element.count();
      return {
        success: true,
        data: {
          selector: command.selector,
          wait_condition: command.waitCondition || 'visible',
          waited_ms: waitedMs,
          element_found: count > 0,
        },
      };
    }

    case 'query-all': {
      const locator = page.locator(command.selector);
      const count = await locator.count();
      if (count === 0) {
        throw new Error(`No elements found: ${command.selector}`);
      }

      const limit =
        typeof command.limit === 'number' && command.limit > 0 ? Math.min(command.limit, count) : count;
      const elements = [];
      const shouldTrim = command.trim !== false;

      for (let i = 0; i < limit; i += 1) {
        const elementHandle = locator.nth(i);
        const item = { index: i };

        if (command.getText) {
          let text = await elementHandle.textContent({ timeout });
          if (shouldTrim && text) {
            text = text.replace(/^\s+|\s+$/g, '').replace(/\s+/g, ' ');
          }
          item.text = text;
        }

        if (command.attributeName) {
          item.attributes = {
            [command.attributeName]: await elementHandle.getAttribute(command.attributeName),
          };
        }

        elements.push(item);
      }

      return {
        success: true,
        data: {
          selector: command.selector,
          element_count: count,
          limit,
          elements,
        },
      };
    }

    case 'get-html': {
      let html;
      if (command.selector) {
        const element = page.locator(command.selector);
        const count = await element.count();
        if (count === 0) {
          throw new Error(`Element not found: ${command.selector}`);
        }
        html = await element.innerHTML({ timeout });
      } else {
        html = await page.content();
      }

      return {
        success: true,
        data: {
          html,
          html_length: html.length,
          selector: command.selector || null,
        },
      };
    }

    case 'evaluate': {
      try {
        const result = await page.evaluate(command.script);
        let resultType = typeof result;
        if (result === null) {
          resultType = 'null';
        } else if (Array.isArray(result)) {
          resultType = 'array';
        }

        return {
          success: true,
          data: {
            result,
            result_type: resultType,
          },
        };
      } catch (error) {
        return {
          success: false,
          error: `Script execution failed: ${error.message}`,
        };
      }
    }

    case 'ping':
      return {
        success: true,
        data: { status: 'alive' },
      };

    default:
      return {
        success: false,
        error: `Unknown command: ${command.command}`,
      };
  }
}

(async () => {
  const { browserType, headless } = parseConfig();
  const launcher = resolveBrowserLauncher(browserType);
  const browser = await launcher.launch({ headless });
  const page = await browser.newPage();
  const version = await browser.version();
  const isConnected = browser.isConnected();

  const server = net.createServer((socket) => {
    let buffer = '';

    socket.on('data', async (data) => {
      buffer += data.toString();

      const lines = buffer.split('\n');
      buffer = lines.pop();

      for (const line of lines) {
        if (!line.trim()) {
          continue;
        }

        let command;
        try {
          command = JSON.parse(line);
        } catch (error) {
          socket.write(`${toCommandError(error)}\n`);
          continue;
        }

        try {
          const response = await handleCommand(page, command);
          socket.write(`${JSON.stringify(response)}\n`);
        } catch (error) {
          socket.write(`${toCommandError(error)}\n`);
        }
      }
    });

    socket.on('error', () => {
      // Ignore socket errors; Go side will handle retries/cleanup.
    });
  });

  server.on('error', (error) => {
    console.log(
      JSON.stringify({
        success: false,
        error: error.message,
      }),
    );
    process.exit(1);
  });

  server.listen(0, '127.0.0.1', () => {
    const { port } = server.address();

    console.log(
      JSON.stringify({
        success: true,
        data: {
          browserType,
          headless,
          version,
          isConnected,
          port,
        },
      }),
    );
  });

  const shutdown = async () => {
    server.close();
    await browser.close();
    process.exit(0);
  };

  process.on('SIGINT', shutdown);
  process.on('SIGTERM', shutdown);
})().catch((error) => {
  console.log(
    JSON.stringify({
      success: false,
      error: error.message,
    }),
  );
  process.exit(1);
});
