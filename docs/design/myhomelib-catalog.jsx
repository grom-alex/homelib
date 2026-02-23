import { useState, useMemo } from "react";

const AUTHORS = [
  { id: 1, name: "Азимов, Айзек", books: 3 },
  { id: 2, name: "Брэдбери, Рэй", books: 2 },
  { id: 3, name: "Гоголь, Николай", books: 2 },
  { id: 4, name: "Достоевский, Фёдор", books: 3 },
  { id: 5, name: "Лем, Станислав", books: 2 },
  { id: 6, name: "Стругацкий, Аркадий", books: 2 },
  { id: 7, name: "Стругацкий, Борис", books: 2 },
  { id: 8, name: "Толкин, Дж. Р. Р.", books: 3 },
  { id: 9, name: "Толстой, Лев", books: 2 },
  { id: 10, name: "Чехов, Антон", books: 2 },
];

const SERIES = [
  { id: 1, name: "Основание", author: "Азимов, Айзек", books: 3 },
  { id: 2, name: "Марсианские хроники", author: "Брэдбери, Рэй", books: 1 },
  { id: 3, name: "Властелин колец", author: "Толкин, Дж. Р. Р.", books: 3 },
  { id: 4, name: "Кибериада", author: "Лем, Станислав", books: 1 },
  { id: 5, name: "Полдень, XXII век", author: "Стругацкие", books: 2 },
  { id: 6, name: "Без серии", author: "", books: 8 },
];

const GENRES = [
  {
    id: 1, name: "Проза", children: [
      { id: 11, name: "Классическая проза" },
      { id: 12, name: "Современная проза" },
      { id: 13, name: "Историческая проза" },
      { id: 14, name: "Русская классика" },
    ]
  },
  {
    id: 2, name: "Фантастика", children: [
      { id: 21, name: "Научная фантастика" },
      { id: 22, name: "Космическая фантастика" },
      { id: 23, name: "Социальная фантастика" },
      { id: 24, name: "Киберпанк" },
    ]
  },
  {
    id: 3, name: "Фэнтези", children: [
      { id: 31, name: "Эпическое фэнтези" },
      { id: 32, name: "Тёмное фэнтези" },
      { id: 33, name: "Городское фэнтези" },
    ]
  },
  {
    id: 4, name: "Детективы", children: [
      { id: 41, name: "Классический детектив" },
      { id: 42, name: "Триллер" },
      { id: 43, name: "Исторический детектив" },
    ]
  },
  {
    id: 5, name: "Поэзия", children: [
      { id: 51, name: "Классическая поэзия" },
      { id: 52, name: "Современная поэзия" },
    ]
  },
];

const ALL_BOOKS = [
  { id: 1, title: "Основание", author: "Азимов, Айзек", series: "Основание", seriesNo: 1, genre: "Научная фантастика", size: "412 КБ", format: "fb2", rating: 5, year: 1951, lang: "ru", annotation: "Первый роман знаменитого цикла Айзека Азимова. Математик Гэри Селдон предсказывает падение Галактической Империи и создаёт план по сокращению грядущего периода варварства с тридцати тысяч лет до одного тысячелетия." },
  { id: 2, title: "Основание и Империя", author: "Азимов, Айзек", series: "Основание", seriesNo: 2, genre: "Научная фантастика", size: "386 КБ", format: "fb2", rating: 5, year: 1952, lang: "ru", annotation: "Второй роман цикла. Основание сталкивается с угрозой, которую не предвидел даже Гэри Селдон — загадочный мутант Мул, способный контролировать эмоции людей." },
  { id: 3, title: "Второе Основание", author: "Азимов, Айзек", series: "Основание", seriesNo: 3, genre: "Научная фантастика", size: "354 КБ", format: "fb2", rating: 5, year: 1953, lang: "ru", annotation: "Третий роман цикла. Поиски таинственного Второго Основания, скрытого где-то на другом конце Галактики." },
  { id: 4, title: "451 градус по Фаренгейту", author: "Брэдбери, Рэй", series: "", seriesNo: 0, genre: "Научная фантастика", size: "198 КБ", format: "fb2", rating: 5, year: 1953, lang: "ru", annotation: "Антиутопический роман о мире, где пожарные не тушат пожары, а сжигают книги. Гай Монтэг — пожарный, который начинает сомневаться в своей миссии." },
  { id: 5, title: "Марсианские хроники", author: "Брэдбери, Рэй", series: "Марсианские хроники", seriesNo: 1, genre: "Научная фантастика", size: "287 КБ", format: "epub", rating: 5, year: 1950, lang: "ru", annotation: "Сборник рассказов о колонизации Марса людьми, о встрече с марсианской цивилизацией и о судьбе человечества." },
  { id: 6, title: "Мёртвые души", author: "Гоголь, Николай", series: "", seriesNo: 0, genre: "Русская классика", size: "542 КБ", format: "fb2", rating: 5, year: 1842, lang: "ru", annotation: "Поэма в прозе о похождениях Павла Ивановича Чичикова, скупающего «мёртвые души» — записи об умерших крепостных крестьянах." },
  { id: 7, title: "Ревизор", author: "Гоголь, Николай", series: "", seriesNo: 0, genre: "Русская классика", size: "156 КБ", format: "fb2", rating: 5, year: 1836, lang: "ru", annotation: "Комедия о мелком чиновнике Хлестакове, которого жители уездного города принимают за ревизора из Петербурга." },
  { id: 8, title: "Преступление и наказание", author: "Достоевский, Фёдор", series: "", seriesNo: 0, genre: "Русская классика", size: "723 КБ", format: "fb2", rating: 5, year: 1866, lang: "ru", annotation: "Роман о бывшем студенте Родионе Раскольникове, совершившем убийство и терзаемом муками совести." },
  { id: 9, title: "Идиот", author: "Достоевский, Фёдор", series: "", seriesNo: 0, genre: "Русская классика", size: "812 КБ", format: "fb2", rating: 4, year: 1869, lang: "ru", annotation: "Роман о князе Мышкине — «положительно прекрасном человеке», столкнувшемся с жестокостью и цинизмом петербургского общества." },
  { id: 10, title: "Братья Карамазовы", author: "Достоевский, Фёдор", series: "", seriesNo: 0, genre: "Русская классика", size: "1.1 МБ", format: "fb2", rating: 5, year: 1880, lang: "ru", annotation: "Последний роман Достоевского. Философская драма о трёх братьях — Дмитрии, Иване и Алёше Карамазовых и их отце Фёдоре Павловиче." },
  { id: 11, title: "Солярис", author: "Лем, Станислав", series: "", seriesNo: 0, genre: "Научная фантастика", size: "298 КБ", format: "fb2", rating: 5, year: 1961, lang: "ru", annotation: "Роман о контакте с внеземным разумом — мыслящим океаном планеты Солярис, который материализует подавленные воспоминания членов экипажа станции." },
  { id: 12, title: "Кибериада", author: "Лем, Станислав", series: "Кибериада", seriesNo: 1, genre: "Научная фантастика", size: "445 КБ", format: "epub", rating: 5, year: 1965, lang: "ru", annotation: "Цикл сатирических рассказов о приключениях конструкторов Трурля и Клапауция в мире разумных роботов." },
  { id: 13, title: "Пикник на обочине", author: "Стругацкий, Аркадий", series: "Полдень, XXII век", seriesNo: 0, genre: "Социальная фантастика", size: "234 КБ", format: "fb2", rating: 5, year: 1972, lang: "ru", annotation: "Повесть о сталкере Рэдрике Шухарте, который проникает в Зону — место посещения Земли пришельцами — в поисках артефактов." },
  { id: 14, title: "Трудно быть богом", author: "Стругацкий, Борис", series: "Полдень, XXII век", seriesNo: 0, genre: "Социальная фантастика", size: "267 КБ", format: "fb2", rating: 5, year: 1964, lang: "ru", annotation: "Повесть о земном наблюдателе доне Румате, который живёт на средневековой планете и вынужден наблюдать за торжеством невежества и жестокости." },
  { id: 15, title: "Братство Кольца", author: "Толкин, Дж. Р. Р.", series: "Властелин колец", seriesNo: 1, genre: "Эпическое фэнтези", size: "678 КБ", format: "fb2", rating: 5, year: 1954, lang: "ru", annotation: "Первый том трилогии. Хоббит Фродо Бэггинс получает в наследство Кольцо Всевластья и отправляется в путешествие, чтобы уничтожить его." },
  { id: 16, title: "Две крепости", author: "Толкин, Дж. Р. Р.", series: "Властелин колец", seriesNo: 2, genre: "Эпическое фэнтези", size: "612 КБ", format: "fb2", rating: 5, year: 1954, lang: "ru", annotation: "Второй том трилогии. Братство распалось, и каждый из его членов идёт своим путём в борьбе с тёмным властелином Сауроном." },
  { id: 17, title: "Возвращение короля", author: "Толкин, Дж. Р. Р.", series: "Властелин колец", seriesNo: 3, genre: "Эпическое фэнтези", size: "589 КБ", format: "fb2", rating: 5, year: 1955, lang: "ru", annotation: "Заключительный том трилогии. Решающая битва за Средиземье и поход Фродо к Ородруину." },
  { id: 18, title: "Война и мир", author: "Толстой, Лев", series: "", seriesNo: 0, genre: "Историческая проза", size: "3.2 МБ", format: "fb2", rating: 5, year: 1869, lang: "ru", annotation: "Эпопея о судьбах нескольких семей русского дворянства на фоне Наполеоновских войн. Одно из величайших произведений мировой литературы." },
  { id: 19, title: "Анна Каренина", author: "Толстой, Лев", series: "", seriesNo: 0, genre: "Классическая проза", size: "1.4 МБ", format: "fb2", rating: 5, year: 1877, lang: "ru", annotation: "Роман о трагической любви замужней женщины Анны Карениной и блестящего офицера графа Вронского на фоне жизни русского общества 1870-х годов." },
  { id: 20, title: "Вишнёвый сад", author: "Чехов, Антон", series: "", seriesNo: 0, genre: "Русская классика", size: "87 КБ", format: "fb2", rating: 4, year: 1904, lang: "ru", annotation: "Последняя пьеса Чехова о разорившемся дворянском семействе, вынужденном продать своё родовое имение с вишнёвым садом." },
  { id: 21, title: "Три сестры", author: "Чехов, Антон", series: "", seriesNo: 0, genre: "Русская классика", size: "112 КБ", format: "fb2", rating: 4, year: 1901, lang: "ru", annotation: "Пьеса о трёх сёстрах Прозоровых, мечтающих вернуться в Москву из провинциального города и тщетно ожидающих перемен в жизни." },
];

const StarRating = ({ rating }) => (
  <span style={{ color: "#d4a017", letterSpacing: 1, fontSize: 12 }}>
    {"★".repeat(rating)}{"☆".repeat(5 - rating)}
  </span>
);

const SearchIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"><circle cx="11" cy="11" r="7"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
);

const BookIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
);

const FolderIcon = ({ open }) => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill={open ? "#d4a017" : "none"} stroke="currentColor" strokeWidth="2">
    {open
      ? <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
      : <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
    }
  </svg>
);

const ChevronIcon = ({ open }) => (
  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"
    style={{ transform: open ? "rotate(90deg)" : "rotate(0deg)", transition: "transform 0.15s" }}>
    <polyline points="9 18 15 12 9 6"/>
  </svg>
);

const UserIcon = () => (
  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
);

const ListIcon = () => (
  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
);

const TagIcon = () => (
  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg>
);

const css = `
  @import url('https://fonts.googleapis.com/css2?family=Source+Sans+3:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap');

  * { margin: 0; padding: 0; box-sizing: border-box; }

  :root {
    --bg-main: #1a1d23;
    --bg-panel: #21252b;
    --bg-header: #181b20;
    --bg-row-hover: #2a2e36;
    --bg-row-selected: #2c3340;
    --bg-input: #181b20;
    --border: #2e333b;
    --border-light: #363b44;
    --text-primary: #c8cdd5;
    --text-secondary: #8b919a;
    --text-muted: #5c6370;
    --accent: #d4a017;
    --accent-dim: rgba(212,160,23,0.12);
    --accent-hover: #e6b422;
    --nav-active: #d4a017;
    --scrollbar-bg: #21252b;
    --scrollbar-thumb: #3a3f48;
  }

  body {
    font-family: 'Source Sans 3', sans-serif;
    background: var(--bg-main);
    color: var(--text-primary);
    overflow: hidden;
    height: 100vh;
  }

  #root { height: 100vh; display: flex; flex-direction: column; }

  ::-webkit-scrollbar { width: 8px; height: 8px; }
  ::-webkit-scrollbar-track { background: var(--scrollbar-bg); }
  ::-webkit-scrollbar-thumb { background: var(--scrollbar-thumb); border-radius: 4px; }
  ::-webkit-scrollbar-thumb:hover { background: #4a5060; }

  .resizer {
    width: 4px;
    cursor: col-resize;
    background: var(--border);
    transition: background 0.2s;
    flex-shrink: 0;
  }
  .resizer:hover, .resizer.active { background: var(--accent); }

  .resizer-h {
    height: 4px;
    cursor: row-resize;
    background: var(--border);
    transition: background 0.2s;
    flex-shrink: 0;
  }
  .resizer-h:hover, .resizer-h.active { background: var(--accent); }

  .genre-item { cursor: pointer; padding: 3px 4px; border-radius: 3px; display: flex; align-items: center; gap: 4px; user-select: none; }
  .genre-item:hover { background: var(--bg-row-hover); }
  .genre-item.selected { background: var(--accent-dim); color: var(--accent); }

  .genre-child { cursor: pointer; padding: 3px 4px 3px 22px; border-radius: 3px; display: flex; align-items: center; gap: 5px; user-select: none; font-size: 13px; }
  .genre-child:hover { background: var(--bg-row-hover); }
  .genre-child.selected { background: var(--accent-dim); color: var(--accent); }

  .list-item { cursor: pointer; padding: 5px 10px; border-bottom: 1px solid var(--border); display: flex; justify-content: space-between; align-items: center; font-size: 13px; }
  .list-item:hover { background: var(--bg-row-hover); }
  .list-item.selected { background: var(--accent-dim); color: var(--accent); }

  .book-row { cursor: pointer; display: flex; border-bottom: 1px solid var(--border); font-size: 13px; }
  .book-row:hover { background: var(--bg-row-hover); }
  .book-row.selected { background: var(--bg-row-selected); }

  .book-cell { padding: 6px 10px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; border-right: 1px solid var(--border); flex-shrink: 0; }

  .search-field { display: flex; flex-direction: column; gap: 4px; }
  .search-field label { font-size: 11px; text-transform: uppercase; letter-spacing: 0.5px; color: var(--text-muted); font-weight: 600; }
  .search-field input, .search-field select {
    background: var(--bg-input); border: 1px solid var(--border); color: var(--text-primary);
    padding: 6px 10px; border-radius: 4px; font-size: 13px; font-family: inherit;
    outline: none; transition: border-color 0.2s;
  }
  .search-field input:focus, .search-field select:focus { border-color: var(--accent); }
  .search-field select option { background: var(--bg-panel); }

  @keyframes fadeIn { from { opacity: 0; transform: translateY(4px); } to { opacity: 1; transform: translateY(0); } }
  .fade-in { animation: fadeIn 0.25s ease; }

  @keyframes dropIn { from { opacity: 0; transform: translateY(-6px) scale(0.97); } to { opacity: 1; transform: translateY(0) scale(1); } }

  .user-menu-btn {
    display: flex; align-items: center; gap: 8px; padding: 4px 8px 4px 4px;
    background: none; border: 1px solid transparent; border-radius: 6px;
    cursor: pointer; color: var(--text-secondary); font-family: inherit;
    transition: all 0.15s; position: relative;
  }
  .user-menu-btn:hover { background: var(--bg-row-hover); border-color: var(--border); color: var(--text-primary); }
  .user-menu-btn.open { background: var(--bg-row-hover); border-color: var(--border-light); color: var(--text-primary); }

  .user-avatar {
    width: 28px; height: 28px; border-radius: 50%; background: linear-gradient(135deg, #d4a017 0%, #b8860b 100%);
    display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700;
    color: #1a1d23; flex-shrink: 0; letter-spacing: -0.5px;
  }

  .user-dropdown {
    position: absolute; top: calc(100% + 6px); right: 0; z-index: 100;
    background: var(--bg-panel); border: 1px solid var(--border-light);
    border-radius: 8px; min-width: 220px; box-shadow: 0 12px 32px rgba(0,0,0,0.4);
    animation: dropIn 0.18s ease; overflow: hidden;
  }

  .user-dropdown-header {
    padding: 14px 14px 12px; border-bottom: 1px solid var(--border);
    display: flex; align-items: center; gap: 10px;
  }

  .user-dropdown-item {
    display: flex; align-items: center; gap: 10px; padding: 9px 14px;
    font-size: 13px; color: var(--text-secondary); cursor: pointer;
    transition: all 0.12s; border: none; background: none; width: 100%;
    font-family: inherit; text-align: left;
  }
  .user-dropdown-item:hover { background: var(--bg-row-hover); color: var(--text-primary); }
  .user-dropdown-item.danger:hover { background: rgba(220,60,60,0.1); color: #e05555; }

  .user-dropdown-divider { height: 1px; background: var(--border); margin: 4px 0; }

  .user-menu-overlay { position: fixed; inset: 0; z-index: 99; }

  .detail-btn {
    display: flex; align-items: center; gap: 7px; padding: 8px 18px;
    border-radius: 5px; cursor: pointer; font-family: inherit; font-size: 13px;
    font-weight: 600; transition: all 0.15s; border: none; white-space: nowrap;
  }
  .detail-btn.primary {
    background: var(--accent); color: #1a1d23;
  }
  .detail-btn.primary:hover { background: var(--accent-hover); box-shadow: 0 2px 12px rgba(212,160,23,0.25); }
  .detail-btn.secondary {
    background: transparent; color: var(--text-secondary);
    border: 1px solid var(--border-light);
  }
  .detail-btn.secondary:hover { border-color: var(--text-muted); color: var(--text-primary); background: var(--bg-row-hover); }
`;

export default function MyHomeLib() {
  const [activeTab, setActiveTab] = useState("authors");
  const [selectedAuthor, setSelectedAuthor] = useState(null);
  const [selectedSeries, setSelectedSeries] = useState(null);
  const [selectedGenre, setSelectedGenre] = useState(null);
  const [selectedBook, setSelectedBook] = useState(null);
  const [authorSearch, setAuthorSearch] = useState("");
  const [seriesSearch, setSeriesSearch] = useState("");
  const [expandedGenres, setExpandedGenres] = useState(new Set([1, 2]));
  const [leftWidth, setLeftWidth] = useState(280);
  const [topHeight, setTopHeight] = useState(null);
  const [sortCol, setSortCol] = useState("title");
  const [sortDir, setSortDir] = useState("asc");

  // Search tab state
  const [searchTitle, setSearchTitle] = useState("");
  const [searchAuthor, setSearchAuthor] = useState("");
  const [searchSeries, setSearchSeries] = useState("");
  const [searchGenre, setSearchGenre] = useState("");
  const [searchFormat, setSearchFormat] = useState("");
  const [searchResults, setSearchResults] = useState(null);
  const [userMenuOpen, setUserMenuOpen] = useState(false);

  const tabs = [
    { id: "authors", label: "Авторы", icon: <UserIcon /> },
    { id: "series", label: "Серии", icon: <ListIcon /> },
    { id: "genres", label: "Жанры", icon: <TagIcon /> },
    { id: "search", label: "Поиск", icon: <SearchIcon /> },
  ];

  const filteredAuthors = AUTHORS.filter(a =>
    a.name.toLowerCase().includes(authorSearch.toLowerCase())
  );
  const filteredSeries = SERIES.filter(s =>
    s.name.toLowerCase().includes(seriesSearch.toLowerCase())
  );

  const currentBooks = useMemo(() => {
    let books = [];
    if (activeTab === "authors" && selectedAuthor) {
      books = ALL_BOOKS.filter(b => b.author === selectedAuthor);
    } else if (activeTab === "series" && selectedSeries) {
      books = selectedSeries === "Без серии"
        ? ALL_BOOKS.filter(b => !b.series)
        : ALL_BOOKS.filter(b => b.series === selectedSeries);
    } else if (activeTab === "genres" && selectedGenre) {
      books = ALL_BOOKS.filter(b => b.genre === selectedGenre);
    } else if (activeTab === "search" && searchResults) {
      books = searchResults;
    } else if (activeTab === "authors" || activeTab === "series" || activeTab === "genres") {
      books = [];
    }
    const dir = sortDir === "asc" ? 1 : -1;
    return [...books].sort((a, b) => {
      const va = a[sortCol] ?? "";
      const vb = b[sortCol] ?? "";
      if (typeof va === "number") return (va - vb) * dir;
      return String(va).localeCompare(String(vb), "ru") * dir;
    });
  }, [activeTab, selectedAuthor, selectedSeries, selectedGenre, searchResults, sortCol, sortDir]);

  const handleSort = (col) => {
    if (sortCol === col) setSortDir(d => d === "asc" ? "desc" : "asc");
    else { setSortCol(col); setSortDir("asc"); }
  };

  const doSearch = () => {
    const results = ALL_BOOKS.filter(b => {
      if (searchTitle && !b.title.toLowerCase().includes(searchTitle.toLowerCase())) return false;
      if (searchAuthor && !b.author.toLowerCase().includes(searchAuthor.toLowerCase())) return false;
      if (searchSeries && !b.series.toLowerCase().includes(searchSeries.toLowerCase())) return false;
      if (searchGenre && b.genre !== searchGenre) return false;
      if (searchFormat && b.format !== searchFormat) return false;
      return true;
    });
    setSearchResults(results);
    setSelectedBook(null);
  };

  const handleTabChange = (tabId) => {
    setActiveTab(tabId);
    setSelectedBook(null);
    if (tabId !== "search") setSearchResults(null);
  };

  // Resizer logic
  const startResizeX = (e) => {
    e.preventDefault();
    const startX = e.clientX;
    const startW = leftWidth;
    const el = e.target;
    el.classList.add("active");
    const onMove = (ev) => setLeftWidth(Math.max(180, Math.min(500, startW + ev.clientX - startX)));
    const onUp = () => { el.classList.remove("active"); document.removeEventListener("mousemove", onMove); document.removeEventListener("mouseup", onUp); };
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  };

  const startResizeY = (e) => {
    e.preventDefault();
    const startY = e.clientY;
    const container = e.target.parentElement;
    const startH = topHeight || (container.offsetHeight * 0.6);
    const el = e.target;
    el.classList.add("active");
    const onMove = (ev) => setTopHeight(Math.max(120, Math.min(container.offsetHeight - 120, startH + ev.clientY - startY)));
    const onUp = () => { el.classList.remove("active"); document.removeEventListener("mousemove", onMove); document.removeEventListener("mouseup", onUp); };
    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  };

  const toggleGenre = (id) => {
    setExpandedGenres(prev => {
      const n = new Set(prev);
      n.has(id) ? n.delete(id) : n.add(id);
      return n;
    });
  };

  const allGenreNames = GENRES.flatMap(g => g.children.map(c => c.name));

  const columns = [
    { id: "title", label: "Название", width: "35%" },
    { id: "author", label: "Автор", width: "22%" },
    { id: "series", label: "Серия", width: "18%" },
    { id: "genre", label: "Жанр", width: "15%" },
    { id: "size", label: "Размер", width: "10%" },
  ];

  const SortArrow = ({ col }) => {
    if (sortCol !== col) return null;
    return <span style={{ marginLeft: 4, fontSize: 10, opacity: 0.7 }}>{sortDir === "asc" ? "▲" : "▼"}</span>;
  };

  const renderLeftPanel = () => {
    switch (activeTab) {
      case "authors":
        return (
          <div style={{ display: "flex", flexDirection: "column", height: "100%" }} className="fade-in">
            <div style={{ padding: "10px 10px 8px", borderBottom: `1px solid var(--border)` }}>
              <div style={{ position: "relative" }}>
                <input
                  type="text"
                  placeholder="Поиск автора..."
                  value={authorSearch}
                  onChange={e => setAuthorSearch(e.target.value)}
                  style={{
                    width: "100%", padding: "7px 10px 7px 32px", background: "var(--bg-input)",
                    border: "1px solid var(--border)", borderRadius: 4, color: "var(--text-primary)",
                    fontSize: 13, fontFamily: "inherit", outline: "none",
                  }}
                  onFocus={e => e.target.style.borderColor = "var(--accent)"}
                  onBlur={e => e.target.style.borderColor = "var(--border)"}
                />
                <span style={{ position: "absolute", left: 10, top: "50%", transform: "translateY(-50%)", color: "var(--text-muted)" }}><SearchIcon /></span>
              </div>
            </div>
            <div style={{ flex: 1, overflow: "auto" }}>
              {filteredAuthors.map(a => (
                <div key={a.id} className={`list-item${selectedAuthor === a.name ? " selected" : ""}`}
                  onClick={() => { setSelectedAuthor(a.name); setSelectedBook(null); }}>
                  <span>{a.name}</span>
                  <span style={{ fontSize: 11, color: "var(--text-muted)", background: "var(--bg-input)", padding: "1px 7px", borderRadius: 8 }}>{a.books}</span>
                </div>
              ))}
              {filteredAuthors.length === 0 && (
                <div style={{ padding: 20, textAlign: "center", color: "var(--text-muted)", fontSize: 13 }}>Ничего не найдено</div>
              )}
            </div>
          </div>
        );
      case "series":
        return (
          <div style={{ display: "flex", flexDirection: "column", height: "100%" }} className="fade-in">
            <div style={{ padding: "10px 10px 8px", borderBottom: `1px solid var(--border)` }}>
              <div style={{ position: "relative" }}>
                <input
                  type="text"
                  placeholder="Поиск серии..."
                  value={seriesSearch}
                  onChange={e => setSeriesSearch(e.target.value)}
                  style={{
                    width: "100%", padding: "7px 10px 7px 32px", background: "var(--bg-input)",
                    border: "1px solid var(--border)", borderRadius: 4, color: "var(--text-primary)",
                    fontSize: 13, fontFamily: "inherit", outline: "none",
                  }}
                  onFocus={e => e.target.style.borderColor = "var(--accent)"}
                  onBlur={e => e.target.style.borderColor = "var(--border)"}
                />
                <span style={{ position: "absolute", left: 10, top: "50%", transform: "translateY(-50%)", color: "var(--text-muted)" }}><SearchIcon /></span>
              </div>
            </div>
            <div style={{ flex: 1, overflow: "auto" }}>
              {filteredSeries.map(s => (
                <div key={s.id} className={`list-item${selectedSeries === s.name ? " selected" : ""}`}
                  onClick={() => { setSelectedSeries(s.name); setSelectedBook(null); }}>
                  <div>
                    <div>{s.name}</div>
                    {s.author && <div style={{ fontSize: 11, color: "var(--text-muted)", marginTop: 1 }}>{s.author}</div>}
                  </div>
                  <span style={{ fontSize: 11, color: "var(--text-muted)", background: "var(--bg-input)", padding: "1px 7px", borderRadius: 8 }}>{s.books}</span>
                </div>
              ))}
            </div>
          </div>
        );
      case "genres":
        return (
          <div style={{ display: "flex", flexDirection: "column", height: "100%" }} className="fade-in">
            <div style={{ padding: "8px 6px", fontSize: 11, color: "var(--text-muted)", borderBottom: `1px solid var(--border)`, textTransform: "uppercase", letterSpacing: 0.5, fontWeight: 600 }}>
              Дерево жанров
            </div>
            <div style={{ flex: 1, overflow: "auto", padding: "6px" }}>
              {GENRES.map(g => {
                const isOpen = expandedGenres.has(g.id);
                return (
                  <div key={g.id}>
                    <div className="genre-item" onClick={() => toggleGenre(g.id)}>
                      <ChevronIcon open={isOpen} />
                      <FolderIcon open={isOpen} />
                      <span style={{ fontWeight: 500, fontSize: 13 }}>{g.name}</span>
                      <span style={{ fontSize: 11, color: "var(--text-muted)", marginLeft: "auto" }}>{g.children.length}</span>
                    </div>
                    {isOpen && g.children.map(c => (
                      <div key={c.id} className={`genre-child${selectedGenre === c.name ? " selected" : ""}`}
                        onClick={() => { setSelectedGenre(c.name); setSelectedBook(null); }}>
                        <BookIcon />
                        <span>{c.name}</span>
                      </div>
                    ))}
                  </div>
                );
              })}
            </div>
          </div>
        );
      case "search":
        return (
          <div style={{ display: "flex", flexDirection: "column", height: "100%", gap: 0 }} className="fade-in">
            <div style={{ padding: "8px 6px", fontSize: 11, color: "var(--text-muted)", borderBottom: `1px solid var(--border)`, textTransform: "uppercase", letterSpacing: 0.5, fontWeight: 600 }}>
              Критерии поиска
            </div>
            <div style={{ flex: 1, overflow: "auto", padding: "12px 10px", display: "flex", flexDirection: "column", gap: 12 }}>
              <div className="search-field">
                <label>Название</label>
                <input value={searchTitle} onChange={e => setSearchTitle(e.target.value)} placeholder="Введите название..." />
              </div>
              <div className="search-field">
                <label>Автор</label>
                <input value={searchAuthor} onChange={e => setSearchAuthor(e.target.value)} placeholder="Введите автора..." />
              </div>
              <div className="search-field">
                <label>Серия</label>
                <input value={searchSeries} onChange={e => setSearchSeries(e.target.value)} placeholder="Введите серию..." />
              </div>
              <div className="search-field">
                <label>Жанр</label>
                <select value={searchGenre} onChange={e => setSearchGenre(e.target.value)}>
                  <option value="">Все жанры</option>
                  {allGenreNames.map(n => <option key={n} value={n}>{n}</option>)}
                </select>
              </div>
              <div className="search-field">
                <label>Формат</label>
                <select value={searchFormat} onChange={e => setSearchFormat(e.target.value)}>
                  <option value="">Все форматы</option>
                  <option value="fb2">fb2</option>
                  <option value="epub">epub</option>
                </select>
              </div>
              <button onClick={doSearch} style={{
                marginTop: 4, padding: "9px 0", background: "var(--accent)", color: "#1a1d23", border: "none",
                borderRadius: 4, cursor: "pointer", fontWeight: 600, fontSize: 13, fontFamily: "inherit",
                transition: "background 0.2s",
              }}
                onMouseEnter={e => e.target.style.background = "var(--accent-hover)"}
                onMouseLeave={e => e.target.style.background = "var(--accent)"}
              >
                <span style={{ display: "flex", alignItems: "center", justifyContent: "center", gap: 6 }}>
                  <SearchIcon /> Найти
                </span>
              </button>
              <button onClick={() => { setSearchTitle(""); setSearchAuthor(""); setSearchSeries(""); setSearchGenre(""); setSearchFormat(""); setSearchResults(null); setSelectedBook(null); }}
                style={{
                  padding: "7px 0", background: "transparent", color: "var(--text-secondary)", border: `1px solid var(--border)`,
                  borderRadius: 4, cursor: "pointer", fontSize: 12, fontFamily: "inherit", transition: "all 0.2s",
                }}
                onMouseEnter={e => { e.target.style.borderColor = "var(--text-muted)"; e.target.style.color = "var(--text-primary)"; }}
                onMouseLeave={e => { e.target.style.borderColor = "var(--border)"; e.target.style.color = "var(--text-secondary)"; }}
              >
                Очистить
              </button>
            </div>
          </div>
        );
      default: return null;
    }
  };

  const bookDetail = selectedBook ? ALL_BOOKS.find(b => b.id === selectedBook) : null;

  return (
    <>
      <style>{css}</style>
      <div style={{ height: "100vh", display: "flex", flexDirection: "column", background: "var(--bg-main)" }}>
        {/* Header */}
        <header style={{
          background: "var(--bg-header)", borderBottom: `1px solid var(--border)`,
          display: "flex", alignItems: "center", padding: "0 16px", height: 48, flexShrink: 0,
        }}>
          <div style={{ display: "flex", alignItems: "center", gap: 8, marginRight: 32 }}>
            <BookIcon />
            <span style={{ fontWeight: 700, fontSize: 15, letterSpacing: -0.3 }}>MyHomeLib</span>
            <span style={{ fontSize: 11, color: "var(--text-muted)", fontWeight: 400, marginLeft: 2 }}>web</span>
          </div>
          <nav style={{ display: "flex", gap: 0, height: "100%" }}>
            {tabs.map(tab => (
              <button key={tab.id} onClick={() => handleTabChange(tab.id)}
                style={{
                  display: "flex", alignItems: "center", gap: 6, padding: "0 18px",
                  background: "none", border: "none", color: activeTab === tab.id ? "var(--nav-active)" : "var(--text-secondary)",
                  cursor: "pointer", fontFamily: "inherit", fontSize: 13, fontWeight: activeTab === tab.id ? 600 : 400,
                  borderBottom: activeTab === tab.id ? "2px solid var(--nav-active)" : "2px solid transparent",
                  transition: "all 0.15s", height: "100%",
                }}
                onMouseEnter={e => { if (activeTab !== tab.id) e.target.style.color = "var(--text-primary)"; }}
                onMouseLeave={e => { if (activeTab !== tab.id) e.target.style.color = "var(--text-secondary)"; }}
              >
                {tab.icon} {tab.label}
              </button>
            ))}
          </nav>
          <div style={{ marginLeft: "auto", display: "flex", alignItems: "center", gap: 14 }}>
            <div style={{ fontSize: 11, color: "var(--text-muted)" }}>
              Книг: <span style={{ color: "var(--accent)", fontWeight: 600 }}>{ALL_BOOKS.length}</span>
            </div>
            <div style={{ width: 1, height: 20, background: "var(--border)" }} />
            <div style={{ position: "relative" }}>
              <button className={`user-menu-btn${userMenuOpen ? " open" : ""}`} onClick={() => setUserMenuOpen(v => !v)}>
                <div className="user-avatar">АИ</div>
                <span style={{ fontSize: 13, fontWeight: 500 }}>Читатель</span>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round"
                  style={{ transition: "transform 0.15s", transform: userMenuOpen ? "rotate(180deg)" : "rotate(0)" }}>
                  <polyline points="6 9 12 15 18 9"/>
                </svg>
              </button>
              {userMenuOpen && (
                <>
                  <div className="user-menu-overlay" onClick={() => setUserMenuOpen(false)} />
                  <div className="user-dropdown">
                    <div className="user-dropdown-header">
                      <div className="user-avatar" style={{ width: 36, height: 36, fontSize: 14 }}>АИ</div>
                      <div>
                        <div style={{ fontWeight: 600, fontSize: 14, color: "var(--text-primary)" }}>Читатель</div>
                        <div style={{ fontSize: 12, color: "var(--text-muted)", marginTop: 1 }}>reader@myhomelib.ru</div>
                      </div>
                    </div>
                    <div style={{ padding: "4px 0" }}>
                      <button className="user-dropdown-item" onClick={() => setUserMenuOpen(false)}>
                        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
                        Мой профиль
                      </button>
                      <button className="user-dropdown-item" onClick={() => setUserMenuOpen(false)}>
                        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
                        Настройки
                      </button>
                      <button className="user-dropdown-item" onClick={() => setUserMenuOpen(false)}>
                        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>
                        Мои коллекции
                      </button>
                      <button className="user-dropdown-item" onClick={() => setUserMenuOpen(false)}>
                        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                        Загрузить книги
                      </button>
                    </div>
                    <div className="user-dropdown-divider" />
                    <div style={{ padding: "4px 0" }}>
                      <button className="user-dropdown-item danger" onClick={() => setUserMenuOpen(false)}>
                        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
                        Выйти
                      </button>
                    </div>
                  </div>
                </>
              )}
            </div>
          </div>
        </header>

        {/* Main content */}
        <div style={{ flex: 1, display: "flex", overflow: "hidden" }}>
          {/* Left panel */}
          <div style={{ width: leftWidth, flexShrink: 0, background: "var(--bg-panel)", borderRight: `1px solid var(--border)`, display: "flex", flexDirection: "column", overflow: "hidden" }}>
            {renderLeftPanel()}
          </div>

          {/* Vertical resizer */}
          <div className="resizer" onMouseDown={startResizeX} />

          {/* Right panel */}
          <div style={{ flex: 1, display: "flex", flexDirection: "column", overflow: "hidden", position: "relative" }}>
            {/* Books table */}
            <div style={{ height: topHeight || "60%", flexShrink: 0, display: "flex", flexDirection: "column", overflow: "hidden" }}>
              {/* Table header */}
              <div style={{ display: "flex", background: "var(--bg-header)", borderBottom: `1px solid var(--border)`, flexShrink: 0 }}>
                {columns.map(col => (
                  <div key={col.id} onClick={() => handleSort(col.id)}
                    style={{
                      width: col.width, padding: "7px 10px", fontSize: 11, fontWeight: 600,
                      textTransform: "uppercase", letterSpacing: 0.5, color: sortCol === col.id ? "var(--accent)" : "var(--text-muted)",
                      cursor: "pointer", borderRight: `1px solid var(--border)`, userSelect: "none",
                      transition: "color 0.15s",
                    }}>
                    {col.label}<SortArrow col={col.id} />
                  </div>
                ))}
              </div>
              {/* Table body */}
              <div style={{ flex: 1, overflow: "auto" }}>
                {currentBooks.length > 0 ? currentBooks.map(book => (
                  <div key={book.id} className={`book-row${selectedBook === book.id ? " selected" : ""}`}
                    onClick={() => setSelectedBook(book.id)}>
                    <div className="book-cell" style={{ width: "35%" }}>{book.title}</div>
                    <div className="book-cell" style={{ width: "22%" }}>{book.author}</div>
                    <div className="book-cell" style={{ width: "18%", color: book.series ? "var(--text-primary)" : "var(--text-muted)", fontStyle: book.series ? "normal" : "italic" }}>
                      {book.series || "—"}
                    </div>
                    <div className="book-cell" style={{ width: "15%" }}>{book.genre}</div>
                    <div className="book-cell" style={{ width: "10%", fontFamily: "'JetBrains Mono', monospace", fontSize: 12, color: "var(--text-secondary)" }}>{book.size}</div>
                  </div>
                )) : (
                  <div style={{ padding: 40, textAlign: "center", color: "var(--text-muted)", fontSize: 13 }}>
                    {activeTab === "search"
                      ? (searchResults ? "Ничего не найдено" : "Задайте критерии и нажмите «Найти»")
                      : "Выберите элемент в левой панели"
                    }
                  </div>
                )}
              </div>
            </div>

            {/* Horizontal resizer */}
            <div className="resizer-h" onMouseDown={startResizeY} />

            {/* Book details */}
            <div style={{ flex: 1, overflow: "auto", background: "var(--bg-panel)", minHeight: 80 }}>
              {bookDetail ? (
                <div style={{ padding: 16 }} className="fade-in">
                  <div style={{ display: "flex", gap: 20, alignItems: "flex-start" }}>
                    {/* Book icon placeholder */}
                    <div style={{
                      width: 80, height: 110, background: "var(--bg-input)", borderRadius: 4,
                      border: `1px solid var(--border)`, display: "flex", alignItems: "center",
                      justifyContent: "center", flexShrink: 0, color: "var(--accent)", fontSize: 28,
                    }}>
                      📖
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <h2 style={{ fontSize: 18, fontWeight: 700, color: "var(--text-primary)", marginBottom: 4, lineHeight: 1.3 }}>
                        {bookDetail.title}
                      </h2>
                      <div style={{ fontSize: 13, color: "var(--text-secondary)", marginBottom: 10 }}>
                        {bookDetail.author}
                      </div>
                      <div style={{ display: "flex", flexWrap: "wrap", gap: "8px 20px", fontSize: 12, marginBottom: 12 }}>
                        <div><span style={{ color: "var(--text-muted)" }}>Серия: </span>{bookDetail.series || "—"}{bookDetail.seriesNo > 0 && ` (#${bookDetail.seriesNo})`}</div>
                        <div><span style={{ color: "var(--text-muted)" }}>Жанр: </span>{bookDetail.genre}</div>
                        <div><span style={{ color: "var(--text-muted)" }}>Год: </span>{bookDetail.year}</div>
                        <div><span style={{ color: "var(--text-muted)" }}>Формат: </span><span style={{ fontFamily: "'JetBrains Mono', monospace" }}>{bookDetail.format}</span></div>
                        <div><span style={{ color: "var(--text-muted)" }}>Размер: </span><span style={{ fontFamily: "'JetBrains Mono', monospace" }}>{bookDetail.size}</span></div>
                        <div><span style={{ color: "var(--text-muted)" }}>Язык: </span>{bookDetail.lang}</div>
                        <div><span style={{ color: "var(--text-muted)" }}>Рейтинг: </span><StarRating rating={bookDetail.rating} /></div>
                      </div>
                      <div style={{ display: "flex", gap: 10, marginBottom: 12 }}>
                        <button
                          className="detail-btn primary"
                          onClick={() => {}}
                        >
                          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
                          Читать
                        </button>
                        <button
                          className="detail-btn secondary"
                          onClick={() => {}}
                        >
                          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                          Скачать <span style={{ fontFamily: "'JetBrains Mono', monospace", fontSize: 11, opacity: 0.7 }}>({bookDetail.format})</span>
                        </button>
                      </div>
                      <div style={{ fontSize: 13, lineHeight: 1.65, color: "var(--text-secondary)", borderTop: `1px solid var(--border)`, paddingTop: 10 }}>
                        <span style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: 0.5, fontWeight: 600, color: "var(--text-muted)", display: "block", marginBottom: 4 }}>Аннотация</span>
                        {bookDetail.annotation}
                      </div>
                    </div>
                  </div>
                </div>
              ) : (
                <div style={{ padding: 40, textAlign: "center", color: "var(--text-muted)", fontSize: 13 }}>
                  Выберите книгу для просмотра подробной информации
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Status bar */}
        <footer style={{
          background: "var(--bg-header)", borderTop: `1px solid var(--border)`,
          padding: "4px 16px", fontSize: 11, color: "var(--text-muted)",
          display: "flex", justifyContent: "space-between", flexShrink: 0,
        }}>
          <span>
            {activeTab === "authors" && selectedAuthor && `Автор: ${selectedAuthor}`}
            {activeTab === "series" && selectedSeries && `Серия: ${selectedSeries}`}
            {activeTab === "genres" && selectedGenre && `Жанр: ${selectedGenre}`}
            {activeTab === "search" && searchResults && `Результаты поиска`}
            {!(
              (activeTab === "authors" && selectedAuthor) ||
              (activeTab === "series" && selectedSeries) ||
              (activeTab === "genres" && selectedGenre) ||
              (activeTab === "search" && searchResults)
            ) && "Готов"}
          </span>
          <span>
            {currentBooks.length > 0 && `Показано книг: ${currentBooks.length}`}
          </span>
        </footer>
      </div>
    </>
  );
}
