export interface Product {
  id: string;
  brand: string;
  name: string;
  description: string;
  originalPrice: number;
  price: number;
  logoSlug: string;
  logoColor: string;
}

export const products: Product[] = [
  {
    id: "steam-10",
    brand: "Steam",
    name: "Steam Wallet",
    description: "Пополнение кошелька $10",
    originalPrice: 10,
    price: 13.9,
    logoSlug: "steam",
    logoColor: "FFFFFF",
  },
  {
    id: "steam-25",
    brand: "Steam",
    name: "Steam Wallet",
    description: "Пополнение кошелька $25",
    originalPrice: 25,
    price: 33.9,
    logoSlug: "steam",
    logoColor: "FFFFFF",
  },
  {
    id: "steam-50",
    brand: "Steam",
    name: "Steam Wallet",
    description: "Пополнение кошелька $50",
    originalPrice: 50,
    price: 66.9,
    logoSlug: "steam",
    logoColor: "FFFFFF",
  },
  {
    id: "psn-10",
    brand: "PlayStation",
    name: "PSN Card",
    description: "Подарочная карта $10",
    originalPrice: 10,
    price: 13.9,
    logoSlug: "playstation",
    logoColor: "003791",
  },
  {
    id: "psn-25",
    brand: "PlayStation",
    name: "PSN Card",
    description: "Подарочная карта $25",
    originalPrice: 25,
    price: 33.9,
    logoSlug: "playstation",
    logoColor: "003791",
  },
  {
    id: "xbox-10",
    brand: "Xbox",
    name: "Xbox Gift Card",
    description: "Подарочная карта $10",
    originalPrice: 10,
    price: 13.9,
    logoSlug: "xbox",
    logoColor: "107C10",
  },
  {
    id: "xbox-25",
    brand: "Xbox",
    name: "Xbox Gift Card",
    description: "Подарочная карта $25",
    originalPrice: 25,
    price: 33.9,
    logoSlug: "xbox",
    logoColor: "107C10",
  },
  {
    id: "nintendo-10",
    brand: "Nintendo",
    name: "eShop Card",
    description: "Подарочная карта $10",
    originalPrice: 10,
    price: 13.9,
    logoSlug: "nintendo",
    logoColor: "E60012",
  },
  {
    id: "spotify-1m",
    brand: "Spotify",
    name: "Spotify Premium",
    description: "Подписка на 1 месяц",
    originalPrice: 9.99,
    price: 13.9,
    logoSlug: "spotify",
    logoColor: "1DB954",
  },
  {
    id: "netflix-1m",
    brand: "Netflix",
    name: "Netflix Standard",
    description: "Подписка на 1 месяц",
    originalPrice: 15.49,
    price: 20.9,
    logoSlug: "netflix",
    logoColor: "E50914",
  },
];

export function getBrandLogoUrl(slug: string, color: string): string {
  return `https://cdn.simpleicons.org/${slug}/${color}`;
}
