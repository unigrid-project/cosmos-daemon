# this script can be run in parallel with testing.py
# it will use the account from testing.yml
# the more threads you give it the faster txps you will get
# python stress_test.py --daemon_path /home/$USER/go/bin/paxd --duration 10 --num_threads 10

import os
import time
from colorama import Fore, init
import pexpect
import threading
import argparse

init(autoreset=True)

LOG_FILE = "stress_test.log"
FROM_ADDRESS = "unigrid192yf94yat7h2sfsrawzh694d477ck7gnylwxf3"
TO_ADDRESS = "unigrid1xeq4qwyhxfukx0xyultta0r882ev86jjs4yvtc"
CHAIN_ID = "your_chain_id"  # Replace with your actual chain ID

# Define a global variable for the transaction count
tx_count = 0
tx_count_lock = threading.Lock()

def send_tokens(daemon_path):
    home_path = os.path.expanduser("~/.pax")
    cmd = (
        f"{daemon_path} tx bank send {FROM_ADDRESS} {TO_ADDRESS} 100000000uugd "
        f"--chain-id={CHAIN_ID} --home={home_path} --fees=0.025uugd --broadcast-mode=sync --yes --output=json"
    )
    try:
        child = pexpect.spawn(cmd)
        
        # Wait for EOF or timeout
        i = child.expect([pexpect.EOF, pexpect.TIMEOUT])
        if i == 0:  # Process ended
            output = child.before.decode()
            if "txhash" in output:
                return True
            else:
                print(f"{Fore.RED}Transaction failed. Output: {output}")
        return False
    except pexpect.exceptions.EOF as e:
        print(f"{Fore.RED}EOF encountered: {str(e)}")
        return False
    except pexpect.exceptions.TIMEOUT as e:
        print(f"{Fore.RED}Timeout encountered: {str(e)}")
        return False
    except Exception as e:
        print(f"{Fore.RED}An error occurred: {str(e)}")
        return False

def stress_test(daemon_path, duration=10):
    start_time = time.time()
    tx_count = 0
    speeds = []

    with open(LOG_FILE, "w") as log:
        while time.time() - start_time < duration:
            if send_tokens(daemon_path):
                tx_count += 1
            current_speed = tx_count / (time.time() - start_time)
            speeds.append(current_speed)
            avg_speed = sum(speeds) / len(speeds)
            slowdown = avg_speed - current_speed

            print(f"{Fore.GREEN}Speed: {current_speed:.2f} tx/s", end=" ")
            print(f"{Fore.YELLOW}Average: {avg_speed:.2f} tx/s", end=" ")
            print(f"{Fore.RED}Slowdown: {slowdown:.2f} tx/s")

            log.write(f"{time.time() - start_time:.2f}s: Speed: {current_speed:.2f} tx/s, Average: {avg_speed:.2f} tx/s, Slowdown: {slowdown:.2f} tx/s\n")

        print(f"\nTest completed. Total transactions: {tx_count}")

def display_metrics(duration):
    start_time = time.time()
    speeds = []
    while time.time() - start_time < duration:
        time.sleep(1)  # Update metrics every second
        current_speed = tx_count / (time.time() - start_time)
        speeds.append(current_speed)
        avg_speed = sum(speeds) / len(speeds)
        slowdown = avg_speed - current_speed

        print(f"{Fore.GREEN}Speed: {current_speed:.2f} tx/s", end=" ")
        print(f"{Fore.YELLOW}Average: {avg_speed:.2f} tx/s", end=" ")
        print(f"{Fore.RED}Slowdown: {slowdown:.2f} tx/s")

def send_tokens_threaded(daemon_path, duration):
    global tx_count
    start_time = time.time()
    while time.time() - start_time < duration:
        if send_tokens(daemon_path):
            with tx_count_lock:
                tx_count += 1

def stress_test_concurrent(daemon_path, duration=10, num_threads=30):
    global tx_count
    tx_count = 0  # Reset the transaction count

    # Start threads for sending transactions
    threads = []
    for _ in range(num_threads):
        t = threading.Thread(target=send_tokens_threaded, args=(daemon_path, duration))
        threads.append(t)
        t.start()

    # Start a separate thread for displaying metrics
    metrics_thread = threading.Thread(target=display_metrics, args=(duration,))
    metrics_thread.start()

    for t in threads:
        t.join()

    metrics_thread.join()

    print(f"\nTest completed. Total transactions: {tx_count}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Stress test for transaction sending.')
    parser.add_argument('--daemon_path', type=str, required=True, help='Path to the paxd daemon executable')
    parser.add_argument('--duration', type=int, default=10, help='Duration of the stress test in seconds')
    parser.add_argument('--num_threads', type=int, default=30, help='Number of concurrent threads')

    args = parser.parse_args()
    stress_test_concurrent(args.daemon_path, args.duration, args.num_threads)
