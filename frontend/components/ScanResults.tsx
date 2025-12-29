type ScanReport = {
  reportTitle: string;
  riskScore: number;
  top3Risks: string[];
  immediateActions: string[];
  remediationPlan: Array<{
    phase: string;
    steps: string[];
  }>;
  preventionMeasures: string[];
};

type ScanResultsProps = {
  report: ScanReport;
};

export default function ScanResults({ report }: ScanResultsProps) {
  const getRiskColor = (score: number) => {
    if (score >= 7) return "text-red-600";
    if (score >= 4) return "text-yellow-600";
    return "text-green-600";
  };

  return (
    <section className="space-y-8 border border-[#2a2a2a] bg-[#0d0d0d] p-6 rounded-none">
      <header className="space-y-3 border-b border-[#2a2a2a] pb-4">
        <h2 className="text-2xl font-semibold tracking-[0.08em] uppercase text-[#e5e5e5]">
          {report.reportTitle}
        </h2>
        <div className="flex items-center gap-4">
          <span className="text-sm tracking-[0.14em] uppercase text-[#9ca3af]">
            Risk Score:
          </span>
          <span className={`text-2xl font-bold ${getRiskColor(report.riskScore)}`}>
            {report.riskScore}/10
          </span>
        </div>
      </header>

      <div className="space-y-6">
        <div className="space-y-3">
          <h3 className="text-sm tracking-[0.16em] uppercase text-[#9ca3af] border-b border-[#2a2a2a] pb-2">
            Top 3 Critical Risks
          </h3>
          <div className="space-y-4">
            {report.top3Risks.map((risk, idx) => (
              <div
                key={idx}
                className="border-l-4 border-red-700 bg-[#0a0a0a] p-4 space-y-2"
              >
                <p className="text-[#e5e5e5] text-sm leading-relaxed font-mono whitespace-pre-wrap">
                  {risk.replace(/\*\*/g, "").replace(/\*/g, "")}
                </p>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-3">
          <h3 className="text-sm tracking-[0.16em] uppercase text-[#9ca3af] border-b border-[#2a2a2a] pb-2">
            Immediate Actions
          </h3>
          <div className="space-y-2">
            {report.immediateActions.map((action, idx) => (
              <div
                key={idx}
                className="border border-[#2a2a2a] bg-[#0a0a0a] p-3"
              >
                <p className="text-[#e5e5e5] text-sm leading-relaxed font-mono">
                  {action.replace(/\*\*/g, "").replace(/\*/g, "")}
                </p>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-3">
          <h3 className="text-sm tracking-[0.16em] uppercase text-[#9ca3af] border-b border-[#2a2a2a] pb-2">
            Remediation Plan
          </h3>
          <div className="space-y-4">
            {report.remediationPlan.map((phase, idx) => (
              <div
                key={idx}
                className="border border-[#2a2a2a] bg-[#0a0a0a] p-4 space-y-3"
              >
                <h4 className="text-sm font-semibold tracking-[0.12em] uppercase text-[#e5e5e5] border-b border-[#2a2a2a] pb-2">
                  {phase.phase}
                </h4>
                <ul className="space-y-2 ml-4">
                  {phase.steps.map((step, stepIdx) => (
                    <li
                      key={stepIdx}
                      className="text-[#9ca3af] text-sm leading-relaxed font-mono list-disc"
                    >
                      <span className="text-[#e5e5e5]">
                        {step.replace(/\*\*/g, "").replace(/\*/g, "")}
                      </span>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-3">
          <h3 className="text-sm tracking-[0.16em] uppercase text-[#9ca3af] border-b border-[#2a2a2a] pb-2">
            Prevention Measures
          </h3>
          <div className="space-y-2">
            {report.preventionMeasures.map((measure, idx) => (
              <div
                key={idx}
                className="border border-[#2a2a2a] bg-[#0a0a0a] p-3"
              >
                <p className="text-[#e5e5e5] text-sm leading-relaxed font-mono">
                  {measure.replace(/\*\*/g, "").replace(/\*/g, "")}
                </p>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

