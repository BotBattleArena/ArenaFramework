import sys
import json
import random

def main():
    # Log to stderr so it doesn't interfere with the protocol
    print("randombot-python: started", file=sys.stderr)

    axis_names = []

    # sys.stdin is an iterator that reads line by line
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue

        try:
            msg = json.loads(line)
            msg_type = msg.get("type")

            if msg_type == "start":
                axes = msg.get("axes", [])
                axis_names = [ax["name"] for ax in axes]
                for ax in axes:
                    print(f"randombot-python: axis registered: {ax['name']} (default: {ax['value']})", file=sys.stderr)

            elif msg_type == "state":
                # Respond with random axis values between -1.0 and 1.0
                response = {
                    "axes": {name: random.uniform(-1.0, 1.0) for name in axis_names}
                }
                # json.dumps + print provides the required NDJSON format (\n)
                # Setting flush=True ensures the data is sent immediately
                print(json.dumps(response), flush=True)

            elif msg_type == "end":
                print("randombot-python: game ended", file=sys.stderr)
                break

        except json.JSONDecodeError as e:
            print(f"randombot-python: decode error: {e}", file=sys.stderr)
        except Exception as e:
            print(f"randombot-python: error: {e}", file=sys.stderr)

if __name__ == "__main__":
    main()
