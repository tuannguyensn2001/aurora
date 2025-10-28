# Aurora Solver

A Python project that integrates with the z3-resolver library for constraint solving.

## Setup

### Using Virtual Environment (Recommended)

1. Create and activate virtual environment:
```bash
# Create virtual environment
python3 -m venv venv

# Activate virtual environment
source venv/bin/activate

# Or use the provided activation script
./activate.sh
```

2. Install dependencies:
```bash
pip install -r requirements.txt
```

Or install the package in development mode:
```bash
pip install -e .
```

3. Deactivate when done:
```bash
deactivate
```

### Alternative Setup (Global Installation)

```bash
pip install -r requirements.txt
```

## Usage

Make sure your virtual environment is activated, then run:
```bash
python main.py
```

## Project Structure

```
aurora/solver/
├── venv/                 # Virtual environment (created after setup)
├── main.py              # Main application file
├── requirements.txt     # Python dependencies
├── setup.py            # Package configuration
├── activate.sh         # Convenience activation script
└── README.md           # This file
```

## Dependencies

- z3-solver: A constraint solver library for Python
