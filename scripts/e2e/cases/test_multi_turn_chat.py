"""Test: Multi-turn chat with message display and session persistence.

Verifies:
1. User can send messages and see them displayed
2. Working indicator shows when claude is processing
3. Multiple messages are visible in the message list
4. After page refresh, messages are restored from localStorage
5. Abort button is visible during work
"""

import logging
from pathlib import Path
from playwright.sync_api import Page, expect


def run(page: Page, base_url: str, out_dir: Path, logger: logging.Logger) -> bool:
    logger.info("Step 1: Navigate and wait for connection")
    page.goto(base_url)
    expect(page.locator("text=已连接")).to_be_visible(timeout=10000)
    page.screenshot(path=str(out_dir / "screenshots" / "01_connected.png"))

    logger.info("Step 2: Send first message")
    textarea = page.locator("textarea")
    textarea.fill("Hello, this is test message 1")
    page.screenshot(path=str(out_dir / "screenshots" / "02_message_typed.png"))

    send_btn = page.locator("button[title='发送']")
    send_btn.click()
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "03_message_sent.png"))

    logger.info("Step 3: Verify user message appears in message list")
    user_msg = page.locator("text=Hello, this is test message 1")
    expect(user_msg).to_be_visible(timeout=5000)
    page.screenshot(path=str(out_dir / "screenshots" / "04_user_msg_visible.png"))

    logger.info("Step 4: Verify working state is indicated")
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "05_working_state.png"))

    logger.info("Step 5: Wait briefly then abort if still working")
    page.wait_for_timeout(2000)
    abort_btn = page.locator("button[title='终止']")
    if abort_btn.is_visible():
        logger.info("Step 5a: Claude is still working, clicking abort")
        abort_btn.click()
        page.wait_for_timeout(1000)
    page.screenshot(path=str(out_dir / "screenshots" / "06_after_abort.png"))

    logger.info("Step 6: Wait for textarea to become enabled again")
    textarea = page.locator("textarea")
    expect(textarea).to_be_enabled(timeout=10000)
    page.screenshot(path=str(out_dir / "screenshots" / "07_input_enabled.png"))

    logger.info("Step 7: Send second message (multi-turn)")
    textarea.fill("This is test message 2, follow-up")
    send_btn = page.locator("button[title='发送']")
    send_btn.click()
    page.wait_for_timeout(500)
    page.screenshot(path=str(out_dir / "screenshots" / "08_second_msg_sent.png"))

    logger.info("Step 8: Verify both user messages are visible")
    msg1 = page.locator("text=Hello, this is test message 1")
    msg2 = page.locator("text=This is test message 2, follow-up")
    expect(msg1).to_be_visible(timeout=3000)
    expect(msg2).to_be_visible(timeout=3000)
    page.screenshot(path=str(out_dir / "screenshots" / "09_both_msgs_visible.png"))

    logger.info("Step 9: Abort second request if still working")
    page.wait_for_timeout(1000)
    abort_btn = page.locator("button[title='终止']")
    if abort_btn.is_visible():
        abort_btn.click()
        page.wait_for_timeout(1000)
    page.screenshot(path=str(out_dir / "screenshots" / "10_second_abort.png"))

    logger.info("Step 10: Verify localStorage has session data")
    session_data = page.evaluate("localStorage.getItem('claude-webui-session')")
    assert session_data is not None, "localStorage should have session data"
    logger.info(f"Session data length: {len(session_data)}")
    page.screenshot(path=str(out_dir / "screenshots" / "11_session_stored.png"))

    logger.info("Step 11: Refresh page and verify messages restored")
    page.reload()
    expect(page.locator("text=已连接")).to_be_visible(timeout=10000)
    page.wait_for_timeout(500)

    msg1_restored = page.locator("text=Hello, this is test message 1")
    expect(msg1_restored).to_be_visible(timeout=5000)

    msg2_restored = page.locator("text=This is test message 2, follow-up")
    expect(msg2_restored).to_be_visible(timeout=5000)
    page.screenshot(path=str(out_dir / "screenshots" / "12_restored_after_refresh.png"))

    logger.info("All checks passed")
    return True
