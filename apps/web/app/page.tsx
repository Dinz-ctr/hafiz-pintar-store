
const categories = [
  "Semua Produk",
  "Buku Edukasi",
  "Alat Belajar",
  "Mainan Edukatif",
];

const products = [
  {
    name: "Produk Edukasi 01",
    category: "Buku Edukasi",
    price: "Rp --",
    emoji: "📚",
    color: "mint",
  },
  {
    name: "Produk Edukasi 02",
    category: "Alat Belajar",
    price: "Rp --",
    emoji: "✏️",
    color: "yellow",
  },
  {
    name: "Produk Edukasi 03",
    category: "Mainan Edukatif",
    price: "Rp --",
    emoji: "🧩",
    color: "pink",
  },
];

export default function Home() {
  return (
    <main>
      <nav className="navbar">
        <a className="brand" href="#">
          <span className="brand-icon">✿</span>
          Hafiz Pintar Store
        </a>

        <div className="nav-links">
          <a href="#beranda">Beranda</a>
          <a href="#produk">Produk</a>
          <a href="#tentang">Tentang Kami</a>
        </div>

        <a className="cart-button" href="#produk">
          🛒 Keranjang
        </a>
      </nav>

      <section className="hero" id="beranda">
        <div className="hero-content">
          <span className="hero-label">
            ✨ BELAJAR JADI LEBIH MENYENANGKAN
          </span>

          <h1>
            Temukan dunia
            <br />
            belajar yang
            <br />
            <span>lebih seru!</span>
          </h1>

          <p>
            Temukan berbagai kebutuhan edukasi pilihan
            untuk menemani proses belajar dan tumbuh
            bersama keluarga.
          </p>

          <a className="primary-button" href="#produk">
            Jelajahi Produk →
          </a>

          <div className="hero-note">
            🌈 Belajar, bermain, dan bertumbuh bersama.
          </div>
        </div>

        <div className="hero-art">
          <div className="art-circle">
            <span className="art-book">📚</span>
            <span className="art-pencil">✏️</span>
            <span className="art-star">⭐</span>
            <span className="art-puzzle">🧩</span>
          </div>
          <div className="floating-note">
            <span>💡</span>
            <div>
              <strong>Belajar Yuk!</strong>
              <small>Temukan hal baru setiap hari</small>
            </div>
          </div>
        </div>
      </section>

      <section className="category-section">
        <div>
          <span className="eyebrow">PILIH SESUAI KEBUTUHAN</span>
          <h2>Kategori Pilihan</h2>
        </div>

        <div className="category-list">
          {categories.map((category) => (
            <button key={category} className="category-chip">
              {category}
            </button>
          ))}
        </div>
      </section>

      <section className="product-section" id="produk">
        <div className="section-heading">
          <div>
            <span className="eyebrow">KOLEKSI KAMI</span>
            <h2>Produk Pilihan ✨</h2>
            <p>Temukan produk yang mendukung kegiatan belajar.</p>
          </div>

          <a href="#produk" className="see-all">
            Lihat semua →
          </a>
        </div>

        <div className="product-grid">
          {products.map((product) => (
            <article className="product-card" key={product.name}>
              <div className={`product-image ${product.color}`}>
                <span>{product.emoji}</span>
              </div>

              <div className="product-info">
                <small>{product.category}</small>
                <h3>{product.name}</h3>
                <strong>{product.price}</strong>
                <button className="add-button">
                  + Lihat Produk
                </button>
              </div>
            </article>
          ))}
        </div>
      </section>

      <footer id="tentang">
        <strong>✿ Hafiz Pintar Store</strong>
        <p>Belajar menyenangkan, masa depan gemilang.</p>
        <small>© 2026 Hafiz Pintar Store</small>
      </footer>
    </main>
  );
}
