# frozen_string_literal: true
require 'matrix'

module Usecase
  class CheckKyouen
    KyouenData = Struct.new(:points, :lineKyouen)

    class << self
      def check(stage)
        stones = to_stones(stage)
        return false if stones.size != 4

        true
      end

      private

      def to_stones(stage)
        stones = []
        size = Math.sqrt(stage.length)
        size.to_i.times do |i|
          size.to_i.times do |j|
            stones.push(Vector[j, i]) if stage[i * size + j] == '1'
          end
        end
        stones
      end

      def kyouen?(stones)
        # p1,p2の垂直二等分線を求める
        l12 = get_midperpendicular(stones[0], stones[1])
        # p2,p3の垂直二等分線を求める
        l23 = get_midperpendicular(stones[1], stones[2])

        # 交点を求める
        intersection123 = get_intersection(l12, l23)
        if intersection123.nil?
          # p1,p2,p3が直線上に存在する場合
          l34 = get_midperpendicular(stones[2], stones[3])
          intersection234 = get_intersection(l23, l34)
          if intersection234.nil?
            # p2,p3,p4が直線状に存在する場合
            return KyouenData.new(stones, true)
          end
        else
          dist1 = get_distance(stones[0], intersection123)
          dist2 = get_distance(stones[3], intersection123)
          if (dist1 - dist2).abs < 0.0000001
            return KyouenData.new(stones, false)
          end
        end
        nil
      end

      def get_distance(p1, p2)
        dist = p1 - p2
        dist.norm
      end

      def get_intersection(l1, l2)
        f1 = l1[1, 0] - l1[0, 0]
        g1 = l1[1, 1] - l1[0, 1]
        f2 = l2[1, 0] - l2[0, 0]
        g2 = l2[1, 1] - l2[0, 1]

        det = (f2 * g1 - f1 * g2).to_f
        return nil if det.zero?

        dx = l2[0, 0] - l1[0, 0]
        dy = l2[0, 1] - l1[0, 1]
        t1 = (f2 * dy - g2 * dx) / det

        Vector[l1[0, 0] + f1 * t1, l1[0, 1] + g1 * t1]
      end

      def get_midperpendicular(p1, p2)
        midpoint = get_midpoint(p1, p2)
        diff = p1 - p2
        gradient = Vector[diff[1], diff[0] * -1]

        Matrix[midpoint.to_a, (midpoint + gradient).to_a]
      end

      def get_midpoint(p1, p2)
        (p1 + p2) / 2.0
      end
    end
  end
end
