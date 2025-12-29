export default function HowItWorks() {
  const steps = [
    "Provide a GitHub repository URL",
    "Select framework & backend stack",
    "Repository is cloned into an isolated workspace",
    "Static + dependency scans are executed",
    "Sandboxed runtime checks are performed",
    "Final security report is generated",
  ];

  return (
    <section className="px-8 py-16 lg:px-16 bg-[#0a0a0a]">
      <div className="max-w-6xl mx-auto space-y-8">
        <div className="flex items-center gap-4">
          <div className="h-px w-12 bg-[#2a2a2a]" />
          <h2 className="text-xl tracking-[0.2em] uppercase">How It Works</h2>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {steps.map((step, idx) => (
            <div
              key={step}
              className="border border-[#2a2a2a] bg-[#111111] px-6 py-6 flex flex-col gap-3 rounded-none"
            >
              <div className="flex items-center justify-between text-sm text-[#9ca3af] tracking-[0.14em] uppercase">
                <span>Step {idx + 1}</span>
                <span className="text-[#e5e5e5]">#{String(idx + 1).padStart(2, "0")}</span>
              </div>
              <p className="text-[#e5e5e5] leading-relaxed">{step}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}