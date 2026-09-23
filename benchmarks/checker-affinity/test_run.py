"""Runner ordering/resumption tests; no Go builds or timing measurements."""

import argparse
import csv
import tempfile
import unittest
from pathlib import Path

from run import Experiment


class RunnerTest(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name).resolve()
        (self.root / "tsconfig.json").write_text("{}")
        (self.root / "index.ts").write_text("export const x = 1;\n")

    def experiment(self, versions):
        args = argparse.Namespace(output=self.root / "results", variant=[(v, v) for v in versions],
                                  project_root=self.root, projects=None, config=Path("tsconfig.json"),
                                  names=["private"], workers=[4], samples=2, scenarios=["full-lint"],
                                  resume_partial=True)
        experiment = Experiment(args)
        self.measured = []

        def invoke(version, case, benchmark=False):
            if not benchmark:
                return ["index.ts"]
            self.assertEqual(case["root"], str(self.root))
            self.assertEqual(case["config"], str(self.root / "tsconfig.json"))
            self.assertEqual(case["files"], ["index.ts"])
            self.measured.append(version)
            return {"ns/op": 1}

        experiment.invoke = invoke
        return experiment

    def read_rows(self):
        with (self.root / "results/samples.csv").open() as stream:
            return list(csv.DictReader(stream))

    def write_rows(self, rows):
        with (self.root / "results/samples.csv").open("w", newline="") as stream:
            writer = csv.DictWriter(stream, fieldnames=list(rows[0]))
            writer.writeheader()
            writer.writerows(rows)

    def test_two_and_four_variant_order_and_resume(self):
        for versions in (["baseline", "candidate"], ["baseline", "affinity", "stealing", "sorted"]):
            with self.subTest(versions=versions):
                (self.root / "results/samples.csv").unlink(missing_ok=True)
                experiment = self.experiment(versions)
                order = [*versions, *reversed(versions)]
                experiment.measurements()
                self.assertEqual(self.measured, order)
                rows = self.read_rows()
                self.write_rows(rows[:3])
                self.measured.clear()
                experiment.measurements()
                self.assertEqual(self.measured, order[3:])
                self.assertEqual([r["version"] for r in self.read_rows()], order)
                self.measured.clear()
                experiment.measurements()
                self.assertEqual(self.measured, [])

    def test_reject_reordered_partial_case(self):
        experiment = self.experiment(["baseline", "affinity", "stealing", "sorted"])
        experiment.measurements()
        rows = self.read_rows()
        self.write_rows([rows[1], rows[0]])
        with self.assertRaisesRegex(RuntimeError, "ordered ABBA prefix"):
            experiment.measurements()

    def test_reject_changed_selected_source(self):
        experiment = self.experiment(["baseline", "candidate"])
        experiment.measurements()
        (self.root / "index.ts").write_text("export const x = 2;\n")
        with self.assertRaisesRegex(RuntimeError, "sources/config changed"):
            experiment.measurements()


if __name__ == "__main__":
    unittest.main()
