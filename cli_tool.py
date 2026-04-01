import argparse
import sys

def main():
    parser = argparse.ArgumentParser(
        description="A simple CLI tool with parameter validation.",
        epilog="Example: python cli_tool.py --name Alice --greet Hello"
    )
    
    parser.add_argument(
        '--name',
        type=str,
        required=True,
        help="The name of the person to greet."
    )
    
    parser.add_argument(
        '--greet',
        type=str,
        required=True,
        help="The greeting message to display."
    )

    try:
        args = parser.parse_args()
    except SystemExit as e:
        # argparse automatically prints a friendly error message and exits
        # when required arguments are missing. We catch the exit to ensure
        # clean handling if needed, though typically we just let it exit.
        sys.exit(e.code)

    # If validation passes (no exception raised), proceed with logic
    print(f"{args.greet}, {args.name}!")

if __name__ == "__main__":
    main()
