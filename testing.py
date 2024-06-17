# to use this script, run the following command:
# python testing.py <path_to_hedgehog_bin> <num_tests>
# this script requires ignite cli to be installed
# python environment should have the following packages installed
# python3 -m venv venv
# source venv/bin/activate
# pip install hdwallet bech32 hashlib binascii urllib3 termcolor requests
# you can monitor what was added to hedgehog in postman or using curl
# https://127.0.0.1:40005/gridspork/mint-storage
# https://127.0.0.1:40005/gridspork/vesting-storage/

import os
import sys
import subprocess
import time
import json
import requests
import urllib3
import hashlib
import bech32
import binascii
import random
import datetime
from datetime import datetime, timezone, timedelta
from mnemonic import Mnemonic
from hdwallet import HDWallet
from hdwallet.symbols import ATOM
from termcolor import colored

# Disable InsecureRequestWarning
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

def execute_command(command, log_file=None):
    print(colored(f"Executing: {command}", "blue"))
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Error: {result.stderr}")
    if log_file:
        with open(log_file, 'a') as f:
            f.write(result.stdout)
            f.write(result.stderr)
    else:
        print(f"Output: {result.stdout.strip()}")
    return result.stdout.strip()

def kill_processes(process_name):
    try:
        result = subprocess.run(["ps", "aux"], capture_output=True, text=True)
        current_pid = os.getpid()
        for line in result.stdout.splitlines():
            if process_name in line and "grep" not in line and str(current_pid) not in line:
                pid = int(line.split()[1])
                print(f"Killing process {pid}")
                subprocess.run(["kill", "-9", str(pid)])
                time.sleep(2)  # Ensure the process has time to terminate
    except Exception as e:
        print(f"No {process_name} process was found to kill. Proceeding...")

def get_json_field(json_string, field):
    data = json.loads(json_string)
    return data.get(field)

def get_current_block_height():
    response = requests.get("http://127.0.0.1:26657/block")
    height = response.json()['result']['block']['header']['height']
    print(colored(f"Current Block Height: {height}", "light_cyan"), end='\r', flush=True)
    return height

def wait_for_ignite_chain(timeout=60):
    start_time = time.time()
    while time.time() - start_time < timeout:
        try:
            response = requests.get("http://127.0.0.1:26657/status")
            if response.status_code == 200:
                print(colored("Ignite chain is up and running.", "green"))
                return True
        except requests.exceptions.ConnectionError:
            pass
        time.sleep(5)
    print(colored("Timeout while waiting for Ignite chain to start.", "red"))
    return False

def wait_for_block_height(target_height, timeout=120):
    start_time = time.time()
    while time.time() - start_time < timeout:
        current_height = int(get_current_block_height())
        if current_height >= target_height:
            return True
        time.sleep(1)
    return False

def verify_mint_storage(rest_port):
    response = requests.get(f"https://127.0.0.1:{rest_port}/gridspork/mint-storage", verify=False)
    return response.json()

def verify_vesting_storage(rest_port):
    response = requests.get(f"https://127.0.0.1:{rest_port}/gridspork/vesting-storage", verify=False)
    return response.json()

def verify_account_vesting(account_address):
    print(colored(f"Verifying account vesting: http://127.0.0.1:1317/cosmos/auth/v1beta1/accounts/{account_address}", "light_magenta"))
    response = requests.get(f"http://127.0.0.1:1317/cosmos/auth/v1beta1/accounts/{account_address}")
    return response.json()

def generate_address_and_keys():
    mnemo = Mnemonic("english")
    seed_phrase = mnemo.generate(strength=256)
    seed = mnemo.to_seed(seed_phrase)
    print(f"Seed: {seed}")
    # Verify seed is a valid hexadecimal string
    try:
        seed_hex = binascii.hexlify(seed).decode()
    except binascii.Error:
        print("Seed is not a valid hexadecimal string")
        return None
    hdwallet = HDWallet(symbol=ATOM)
    hdwallet.from_seed(seed_hex)
    hdwallet.from_path("m/44'/118'/0'/0/0")

    private_key = hdwallet.private_key()
    public_key = hdwallet.public_key()

    # Generate address from public key using SHA256 and RIPEMD160
    pubkey_bytes = bytes.fromhex(public_key)
    sha256 = hashlib.sha256(pubkey_bytes).digest()
    ripemd160 = hashlib.new('ripemd160', sha256).digest()

    # Convert 8-bit binary to 5-bit binary
    converted_bits = bech32.convertbits(ripemd160, 8, 5)
    address = bech32.bech32_encode("unigrid", converted_bits)

    return address, private_key, public_key

def run_test(test_num):
    test_result = {
        "test_num": test_num,
        "status": "pass",
        "steps": [],
        "details": ""
    }
    print(colored(f"Running test {test_num}/{num_tests}", "yellow"))

    wallet_addr, private_key, public_key = generate_address_and_keys()
    print(f"Generated wallet address: {wallet_addr}")
    test_result["steps"].append(f"Generated wallet address: {wallet_addr}")

    # Step 5: Get the current block height
    print(colored("Step 5: Getting the current block height...", "yellow"))
    current_block_height = get_current_block_height()
    test_result["steps"].append(f"Current Block Height: {current_block_height}")

    # Step 6: Generate a random amount for minting
    print(colored("Step 6: Generating a random amount for minting...", "yellow"))
    mint_amount = random.randint(1000000000, 100000000000)  # Example range for mint amount
    print(f"Mint Amount: {mint_amount}")
    test_result["steps"].append(f"Mint Amount: {mint_amount}")

    # Step 7: Submit a mint transaction
    print(colored("Step 7: Submitting a mint transaction...", "yellow"))
    mint_block_height = int(current_block_height) + 24
    mint_url = f"https://127.0.0.1:{rest_port}/gridspork/mint-storage/{wallet_addr}/{mint_block_height}"
    headers = {
        "privateKey": generated_private_key_1,  # Using first generated private key
        "Content-Type": "application/json"
    }
    print(f"Executing Mint Transaction: curl -X PUT '{mint_url}' -H 'privateKey: {generated_private_key_1}' -H 'Content-Type: application/json' -d '{mint_amount}' -k")
    mint_response = requests.put(mint_url, headers=headers, data=str(mint_amount), verify=False)
    print(f"Mint Response Code: {mint_response.status_code}")
    test_result["steps"].append(f"Mint Response Code: {mint_response.status_code}")

    # Step 8: Verify mint storage after a short wait
    print(colored("Step 8: Verifying mint storage after a short wait...", "yellow"))
    time.sleep(5)  # Short wait before checking mint storage
    mint_storage_data = verify_mint_storage(rest_port)
    if not mint_storage_data:
        print(colored("Failed to verify mint storage. Skipping to next test.", "red"))
        test_result["status"] = "fail"
        test_result["details"] = "Failed to verify mint storage."
        test_results.append(test_result)
        return

    # Verify the mint data in Hedgehog
    mint_key = f"{wallet_addr}/{mint_block_height}"
    mint_data_correct = mint_storage_data['data']['mints'].get(mint_key) == mint_amount
    if mint_data_correct:
        print(colored(f"Mint data verification successful: {mint_key} = {mint_amount}", "green"))
        test_result["steps"].append(f"Mint data verification successful: {mint_key} = {mint_amount}")
    else:
        print(colored(f"Mint data verification failed: expected {mint_key} = {mint_amount}, got {mint_storage_data['data']['mints'].get(mint_key)}", "red"))
        test_result["status"] = "fail"
        test_result["details"] = f"Mint data verification failed: expected {mint_key} = {mint_amount}, got {mint_storage_data['data']['mints'].get(mint_key)}"

    # Step 9: Verify balance after the block height is reached
    print(colored("Step 9: Verifying the balance after the block height is reached...", "yellow"))
    balance_url = f"http://127.0.0.1:1317/cosmos/bank/v1beta1/balances/{wallet_addr}"
    balance_data = None
    if not wait_for_block_height(mint_block_height + 1):
        print(colored("Timeout waiting for the mint block height. Skipping to next test.", "red"))
        test_result["status"] = "fail"
        test_result["details"] = "Timeout waiting for the mint block height."
        test_results.append(test_result)
        return
    balance_response = requests.get(balance_url)
    balance_data = balance_response.json()
    balance = balance_data['balances'][0]['amount'] if 'balances' in balance_data and balance_data['balances'] else '0'

    expected_balance = mint_amount
    if int(balance) == expected_balance:
        print(colored(f"Balance verification successful: {balance} uugd", "green"))
        test_result["steps"].append(f"Balance verification successful: {balance} uugd")
    else:
        print(colored(f"Balance verification failed: expected {expected_balance}, got {balance}", "red"))
        test_result["status"] = "fail"
        test_result["details"] = f"Balance verification failed: expected {expected_balance}, got {balance}"

    # Step 10: Get the new block height
    print(colored("Step 10: Getting the current block height...", "yellow"))
    current_block_height = get_current_block_height()
    test_result["steps"].append(f"Current Block Height: {current_block_height}")

    # Step 11: Submit vesting data with the same random amount
    print(colored("Step 11: Submitting vesting data...", "yellow"))
    new_block_height = int(current_block_height) + 25
    start_time = (datetime.now(timezone.utc) + timedelta(minutes=1)).strftime('%Y-%m-%dT%H:%M:%SZ')
    cliff = random.randint(0, 12)  # Random cliff between 0 and 12
    parts = random.randint(6, 36)  # Random parts between 6 and 36
    percent = random.randint(0, 20)  # Random percent between 0 and 20
    vesting_data = {
        "amount": int(mint_amount),
        "block": new_block_height,
        "cliff": cliff,
        "duration": "PT5M",
        "parts": parts,
        "percent": percent,
        "start": start_time
    }
    vesting_url = f"https://127.0.0.1:{rest_port}/gridspork/vesting-storage/{wallet_addr}"
    print(f"Executing Vesting Transaction: curl -X PUT '{vesting_url}' -H 'privateKey: {generated_private_key_2}' -H 'Content-Type: application/json' -d '{json.dumps(vesting_data)}' -k")
    vesting_response = requests.put(vesting_url, headers=headers, json=vesting_data, verify=False)
    print(f"Vesting Response Code: {vesting_response.status_code}")
    print(f"Vesting Response: {vesting_response.text}")
    test_result["steps"].append(f"Vesting Response Code: {vesting_response.status_code}")
    test_result["steps"].append(f"Vesting Response: {vesting_response.text}")

    # Step 12: Verify vesting storage after a short wait
    print(colored("Step 12: Verifying vesting storage after a short wait...", "yellow"))
    time.sleep(5)  # Short wait before checking vesting storage
    vesting_storage_data = verify_vesting_storage(rest_port)
    if not vesting_storage_data:
        print(colored("Failed to verify vesting storage. Skipping to next test.", "red"))
        test_result["status"] = "fail"
        test_result["details"] = "Failed to verify vesting storage."
        test_results.append(test_result)
        return

    # Verify the vesting data in Hedgehog
    vesting_data_correct = (
        vesting_storage_data['data']['vestingAddresses'][f"Address(wif={wallet_addr})"]['amount'] == vesting_data['amount'] and
        vesting_storage_data['data']['vestingAddresses'][f"Address(wif={wallet_addr})"]['block'] == vesting_data['block'] and
        vesting_storage_data['data']['vestingAddresses'][f"Address(wif={wallet_addr})"]['cliff'] == vesting_data['cliff'] and
        vesting_storage_data['data']['vestingAddresses'][f"Address(wif={wallet_addr})"]['duration'] == vesting_data['duration'] and
        vesting_storage_data['data']['vestingAddresses'][f"Address(wif={wallet_addr})"]['parts'] == vesting_data['parts'] and
        vesting_storage_data['data']['vestingAddresses'][f"Address(wif={wallet_addr})"]['percent'] == vesting_data['percent'] and
        vesting_storage_data['data']['vestingAddresses'][f"Address(wif={wallet_addr})"]['start'] == vesting_data['start']
    )
    if vesting_data_correct:
        print(colored(f"Vesting data verification successful: {json.dumps(vesting_data, indent=2)}", "green"))
        test_result["steps"].append(f"Vesting data verification successful: {json.dumps(vesting_data, indent=2)}")
    else:
        print(colored(f"Vesting data verification failed. Expected: {json.dumps(vesting_data, indent=2)}, Got: {json.dumps(vesting_storage_data['data']['vestingAddresses'][f'Address(wif={wallet_addr})'], indent=2)}", "red"))
        test_result["status"] = "fail"
        test_result["details"] = f"Vesting data verification failed. Expected: {json.dumps(vesting_data, indent=2)}, Got: {json.dumps(vesting_storage_data['data']['vestingAddresses'][f'Address(wif={wallet_addr})'], indent=2)}"

    # Step 13: Verify the vesting schedule on the chain
    print(colored("Step 13: Verifying the vesting schedule on the chain...", "yellow"))
    verification_attempts = 0
    while verification_attempts < 3:
        if not wait_for_block_height(new_block_height + 1 + verification_attempts):
            verification_attempts += 1
            continue
        vesting_account_data = verify_account_vesting(wallet_addr)
        vesting_periods = vesting_account_data['account'].get('vesting_periods', [])
        total_vesting_amount = sum(int(period['amount'][0]['amount']) for period in vesting_periods)
        total_parts = len(vesting_periods)

        if vesting_data['cliff'] > 0:
            expected_parts = vesting_data['cliff'] + (vesting_data['parts'] - 1) + 1  # cliff periods + remaining parts + TGE
        else:
            expected_parts = vesting_data['parts']  # parts if no cliff

        if total_vesting_amount == vesting_data['amount'] and total_parts == expected_parts:
            print(colored(f"Vesting periods verification successful: Total Amount = {total_vesting_amount}, Parts = {total_parts}", "green"))
            test_result["steps"].append(f"Vesting periods verification successful: Total Amount = {total_vesting_amount}, Parts = {total_parts}")
            break
        else:
            verification_attempts += 1
            time.sleep(5)  # Wait for 5 seconds before rechecking

    if verification_attempts == 3:
        print(colored(f"Vesting periods verification failed: Expected Amount = {vesting_data['amount']}, Got = {total_vesting_amount}; Expected Parts = {expected_parts}, Got = {total_parts}", "red"))
        test_result["status"] = "fail"
        test_result["details"] = f"Vesting periods verification failed: Expected Amount = {vesting_data['amount']}, Got = {total_vesting_amount}; Expected Parts = {expected_parts}, Got = {total_parts}"

    test_results.append(test_result)

if len(sys.argv) < 3:
    print("Usage: python test_suite.py <path_to_hedgehog_bin> <num_tests>")
    sys.exit(1)

hedgehog_bin = sys.argv[1]
num_tests = int(sys.argv[2])
rest_port = 40005  # Hardcoded Hedgehog URL

# Generate Hedgehog keys (twice)
key_gen_output_1 = execute_command(f'{hedgehog_bin} util key-generate')
lines_1 = key_gen_output_1.splitlines()
generated_private_key_1 = next(line.split(': ')[1] for line in lines_1 if "Private Key" in line)
generated_public_key_1 = next(line.split(': ')[1] for line in lines_1 if "Public Key" in line)

key_gen_output_2 = execute_command(f'{hedgehog_bin} util key-generate')
lines_2 = key_gen_output_2.splitlines()
generated_private_key_2 = next(line.split(': ')[1] for line in lines_2 if "Private Key" in line)
generated_public_key_2 = next(line.split(': ')[1] for line in lines_2 if "Public Key" in line)

# Step 1: Kill any running Hedgehog and Ignite processes
print(colored("Step 1: Killing any running Hedgehog and Ignite processes...", "yellow"))
kill_processes("hedgehog-0.0.8")
kill_processes("ignite")

# Step 2: Remove existing Unigrid local data
print(colored("Step 2: Removing existing Unigrid local data...", "yellow"))
execute_command('rm -rf ~/.local/share/unigrid')

print(colored("Starting the Hedgehog daemon...", "yellow"))
hedgehog_command = f"nohup {hedgehog_bin} daemon --resthost=0.0.0.0 --restport={rest_port} --netport=40002 --no-seeds --network-keys={generated_public_key_1},{generated_public_key_2} -vvvvvv > hedgehog.log 2>&1 &"
execute_command(hedgehog_command)
time.sleep(5)  # Wait for the Hedgehog daemon to start

# Verify if Hedgehog daemon started correctly
try:
    response = requests.get(f"https://127.0.0.1:{rest_port}/gridspork", verify=False)
    if response.status_code == 200:
        print(colored("Hedgehog daemon started successfully.", "green"))
    else:
        print(colored(f"Failed to start Hedgehog daemon: {response.status_code}", "red"))
except Exception as e:
    print(colored(f"Error connecting to Hedgehog daemon: {e}", "red"))
    sys.exit(1)

# Step 3: Use Ignite to set up the chain
print(colored("Step 3: Setting up the chain using Ignite...", "yellow"))
ignite_command = "nohup ignite chain serve --skip-proto --reset-once --config testing.yml -v > ignite.log 2>&1 &"
execute_command(ignite_command, log_file="ignite.log")

# Step 4: Wait for Ignite chain to be up and running
if not wait_for_ignite_chain(timeout=120):
    sys.exit(colored("Failed to start Ignite chain. Exiting.", "red"))

test_results = []

for i in range(num_tests):
    run_test(i + 1)

# Output test results
for result in test_results:
    status = colored("PASS", "green") if result["status"] == "pass" else colored("FAIL", "red")
    print(f"Test {result['test_num']}: {status}")

    if result["status"] == "pass":
        print(colored("Step 14: Killing any running Hedgehog and Ignite processes after all tests...", "yellow"))
        kill_processes("hedgehog-0.0.8")
        kill_processes("ignite")
    if result["status"] == "fail":
        print(f"Details: {result['details']}")
        print("Closing daemon after five minutes to check failed transactions.")
        time.sleep(350)
        kill_processes("hedgehog-0.0.8")
        kill_processes("ignite")
