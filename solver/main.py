from z3 import *

s = Solver()

s.from_string("""
  (declare-const country String)
  (declare-const age Int)
  (declare-const gender String)
  (assert (and (= country "vn") (< age 20) (= gender "male")))
  (assert (and (= country "vn") (< age 16) (= gender "male")))
""")

if s.check() == sat:
    print("Có nghiệm!")
    print(s.model())
else:
    print("Không có nghiệm")