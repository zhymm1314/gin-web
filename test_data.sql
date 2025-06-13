-- 插入测试游戏数据
INSERT INTO games (name, description, created_at, updated_at) VALUES
('Minecraft', 'A sandbox video game developed by Mojang Studios', NOW(), NOW()),
('Skyrim', 'The Elder Scrolls V: Skyrim', NOW(), NOW()),
('GTA V', 'Grand Theft Auto V', NOW(), NOW());

-- 插入测试分类数据
INSERT INTO categories (name, description, created_at, updated_at) VALUES
('Gameplay', 'Mods that change gameplay mechanics', NOW(), NOW()),
('Graphics', 'Visual enhancement mods', NOW(), NOW()),
('Weapons', 'New weapons and armor', NOW(), NOW()),
('Maps', 'New maps and worlds', NOW(), NOW());

-- 插入测试mod数据
INSERT INTO mods (name, description, author, version, game_id, category_id, download_count, rating, download_url, image_url, created_at, updated_at) VALUES
('OptiFine', 'A Minecraft optimization mod that allows for HD textures and many configuration options for better graphics and performance.', 'sp614x', '1.19.4', 1, 2, 15000000, 4.8, 'https://optifine.net/downloads', 'https://optifine.net/img/logo.png', NOW(), NOW()),
('JEI', 'Just Enough Items (JEI) is an item and recipe viewing mod for Minecraft, built from the ground up for stability and performance.', 'mezz', '11.6.0.1018', 1, 1, 8500000, 4.9, 'https://www.curseforge.com/minecraft/mc-mods/jei', '', NOW(), NOW()),
('Biomes O Plenty', 'Adds over 80 unique biomes to enhance your Minecraft world!', 'Forstride', '17.1.2.545', 1, 4, 12000000, 4.7, 'https://www.curseforge.com/minecraft/mc-mods/biomes-o-plenty', '', NOW(), NOW()),
('SkyUI', 'Elegant, PC-friendly interface mod with many advanced features.', 'SkyUI Team', '5.2SE', 2, 2, 3200000, 4.9, 'https://www.nexusmods.com/skyrimspecialedition/mods/12604', '', NOW(), NOW()),
('SKSE64', 'The Skyrim Script Extender (SKSE) is a tool used by many Skyrim mods that expands scripting capabilities.', 'SKSE Team', '2.2.3', 2, 1, 2800000, 4.8, 'https://skse.silverlock.org/', '', NOW(), NOW()),
('Immersive Armors', 'Adds many new armor sets that have been seamlessly integrated into the world.', 'Hothtrooper44', '8.1', 2, 3, 1900000, 4.6, 'https://www.nexusmods.com/skyrimspecialedition/mods/3479', '', NOW(), NOW()),
('NaturalVision Evolved', 'The ultimate GTA V graphics enhancement mod.', 'Razed', '2.0', 3, 2, 850000, 4.7, 'https://www.gta5-mods.com/misc/naturalvision-evolved', '', NOW(), NOW()),
('Script Hook V', 'Library that allows to use GTA V script native functions in custom *.asi plugins.', 'Alexander Blade', '1.0.2845.0', 3, 1, 1200000, 4.5, 'http://www.dev-c.com/gtav/scripthookv/', '', NOW(), NOW()); 