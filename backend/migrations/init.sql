-- =============================================
-- SCHEMA
-- =============================================

CREATE TABLE users (
    id           SERIAL PRIMARY KEY,
    email        VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE states (
    id            SERIAL PRIMARY KEY,
    uf            CHAR(2)        NOT NULL UNIQUE,
    nome          VARCHAR(100)   NOT NULL,
    regiao        VARCHAR(50)    NOT NULL,
    capital       VARCHAR(100)   NOT NULL,
    populacao     BIGINT         NOT NULL,
    pib_per_capita NUMERIC(12,2) NOT NULL
);

CREATE TABLE branches (
    id         SERIAL PRIMARY KEY,
    nome       VARCHAR(255) NOT NULL,
    cidade     VARCHAR(255) NOT NULL,
    state_id   INT          NOT NULL REFERENCES states(id),
    lat        NUMERIC(9,6),
    lng        NUMERIC(9,6),
    opened_at  DATE         NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_branches_state_id  ON branches(state_id);
CREATE INDEX idx_branches_deleted_at ON branches(deleted_at);

-- =============================================
-- SEED: STATES (população Censo 2022, PIB per capita 2021 IBGE)
-- =============================================

INSERT INTO states (uf, nome, regiao, capital, populacao, pib_per_capita) VALUES
('AC', 'Acre',                'Norte',        'Rio Branco',       830026,    17890.45),
('AL', 'Alagoas',             'Nordeste',     'Maceió',          3337357,    14502.30),
('AP', 'Amapá',               'Norte',        'Macapá',           733508,    19234.80),
('AM', 'Amazonas',            'Norte',        'Manaus',          4269995,    24871.60),
('BA', 'Bahia',               'Nordeste',     'Salvador',       14141626,    17453.90),
('CE', 'Ceará',               'Nordeste',     'Fortaleza',       9187103,    17024.50),
('DF', 'Distrito Federal',    'Centro-Oeste', 'Brasília',        2817381,    85661.00),
('ES', 'Espírito Santo',      'Sudeste',      'Vitória',         4064052,    37432.10),
('GO', 'Goiás',               'Centro-Oeste', 'Goiânia',         7113540,    34521.70),
('MA', 'Maranhão',            'Nordeste',     'São Luís',        6775152,    13421.30),
('MT', 'Mato Grosso',         'Centro-Oeste', 'Cuiabá',          3658977,    47832.50),
('MS', 'Mato Grosso do Sul',  'Centro-Oeste', 'Campo Grande',    2833742,    40123.80),
('MG', 'Minas Gerais',        'Sudeste',      'Belo Horizonte', 20732660,    31245.60),
('PA', 'Pará',                'Norte',        'Belém',           8116132,    18932.40),
('PB', 'Paraíba',             'Nordeste',     'João Pessoa',     4039277,    16543.20),
('PR', 'Paraná',              'Sul',          'Curitiba',       11516840,    42318.90),
('PE', 'Pernambuco',          'Nordeste',     'Recife',          9674793,    20341.70),
('PI', 'Piauí',               'Nordeste',     'Teresina',        3270174,    14832.60),
('RJ', 'Rio de Janeiro',      'Sudeste',      'Rio de Janeiro', 16054524,    43258.10),
('RN', 'Rio Grande do Norte', 'Nordeste',     'Natal',           3302490,    19872.30),
('RS', 'Rio Grande do Sul',   'Sul',          'Porto Alegre',   11466630,    43732.80),
('RO', 'Rondônia',            'Norte',        'Porto Velho',     1815278,    27643.50),
('RR', 'Roraima',             'Norte',        'Boa Vista',        636707,    24532.10),
('SC', 'Santa Catarina',      'Sul',          'Florianópolis',   7762154,    49832.60),
('SP', 'São Paulo',           'Sudeste',      'São Paulo',      44411238,    48469.73),
('SE', 'Sergipe',             'Nordeste',     'Aracaju',         2298696,    18234.50),
('TO', 'Tocantins',           'Norte',        'Palmas',          1607363,    25431.80);

-- =============================================
-- SEED: USER
-- =============================================

INSERT INTO users (email, password_hash) VALUES
('admin@empresa.com', '$2a$10$M7Crf0Iv3UUO825CoLVi6efnDUG4iEBDC7B551PpqfKcVR7Tyg2Ie');

-- =============================================
-- SEED: BRANCHES (10 filiais distribuidas)
-- =============================================

INSERT INTO branches (nome, cidade, state_id, lat, lng, opened_at) VALUES
('Filial Centro SP',       'São Paulo',       (SELECT id FROM states WHERE uf='SP'), -23.550520, -46.633308, '2018-03-15'),
('Filial Paulista',        'São Paulo',       (SELECT id FROM states WHERE uf='SP'), -23.561414, -46.655881, '2019-07-22'),
('Filial Barra da Tijuca', 'Rio de Janeiro',  (SELECT id FROM states WHERE uf='RJ'), -23.000385, -43.365894, '2017-11-10'),
('Filial Savassi',         'Belo Horizonte',  (SELECT id FROM states WHERE uf='MG'), -19.939869, -43.938091, '2020-01-08'),
('Filial Asa Norte',       'Brasília',        (SELECT id FROM states WHERE uf='DF'), -15.764799, -47.882781, '2021-05-19'),
('Filial Meireles',        'Fortaleza',       (SELECT id FROM states WHERE uf='CE'), -3.726389,  -38.502220, '2019-09-03'),
('Filial Boa Viagem',      'Recife',          (SELECT id FROM states WHERE uf='PE'), -8.119760,  -34.900681, '2022-02-14'),
('Filial Batel',           'Curitiba',        (SELECT id FROM states WHERE uf='PR'), -25.443800, -49.290000, '2020-08-30'),
('Filial Trindade',        'Florianópolis',   (SELECT id FROM states WHERE uf='SC'), -27.598799, -48.549400, '2021-11-01'),
('Filial Pituba',          'Salvador',        (SELECT id FROM states WHERE uf='BA'), -12.994900, -38.457600, '2023-03-20');