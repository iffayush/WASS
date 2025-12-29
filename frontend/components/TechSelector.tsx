type TechSelectorProps = { value: string; onChange: (value: string) => void };

export default function TechSelector({ value, onChange }: TechSelectorProps) {
  return (
    <div className="space-y-3">
      <h3 className="text-sm tracking-[0.16em] uppercase text-[#9ca3af]">
        Branch
      </h3>
      <div className="grid grid-cols-2 gap-3">
        {["main", "master"].map((branch) => {
          const active = value === branch;
          return (
            <button
              key={branch}
              onClick={() => onChange(branch)}
              className={`px-4 py-3 border-2 rounded-none text-sm tracking-[0.12em] uppercase transition-colors duration-150 ${
                active
                  ? "bg-[#111111] border-red-700 text-[#e5e5e5]"
                  : "bg-[#0a0a0a] border-[#2a2a2a] text-[#9ca3af] hover:text-[#e5e5e5]"
              }`}
            >
              {branch}
            </button>
          );
        })}
      </div>
    </div>
  );
}