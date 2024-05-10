CREATE TABLE IF NOT EXISTS `example`
(
    `id`                       bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `key`                      varchar(64) NOT NULL COMMENT '数据唯一键',
    `data`                     varchar(64) NOT NULL COMMENT '数据内容',
    PRIMARY KEY (`id`) USING BTREE COMMENT '主键索引',
    UNIQUE KEY `uniq_key` (`key`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COMMENT '一致性缓存示例表';