export interface Country {
  code: string;
  name: string;
  planCount: number | null;
}

export const countries: Country[] = [
  { code: "TR", name: "Турция", planCount: 3 },
  { code: "TH", name: "Таиланд", planCount: 3 },
  { code: "US", name: "США", planCount: 3 },
  { code: "AE", name: "ОАЭ", planCount: 3 },
  { code: "JP", name: "Япония", planCount: 2 },
  { code: "ID", name: "Индонезия", planCount: 2 },
  { code: "DE", name: "Германия", planCount: 2 },
  { code: "FR", name: "Франция", planCount: 2 },
  { code: "IT", name: "Италия", planCount: 2 },
  { code: "ES", name: "Испания", planCount: 2 },
  { code: "GB", name: "Великобритания", planCount: 2 },
  { code: "KR", name: "Южная Корея", planCount: 2 },
  { code: "SG", name: "Сингапур", planCount: 2 },
  { code: "MY", name: "Малайзия", planCount: 2 },
  { code: "IN", name: "Индия", planCount: 2 },
  { code: "BR", name: "Бразилия", planCount: 2 },
  { code: "MX", name: "Мексика", planCount: 2 },
  { code: "AU", name: "Австралия", planCount: 2 },
  { code: "EG", name: "Египет", planCount: 2 },
  { code: "GE", name: "Грузия", planCount: null },
];

export function getFlagUrl(code: string): string {
  return `https://flagcdn.com/w80/${code.toLowerCase()}.png`;
}
